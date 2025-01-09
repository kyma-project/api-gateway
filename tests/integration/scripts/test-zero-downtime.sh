#!/bin/bash
# Script to run zero downtime tests by executing a godog integration test for a given handler and sending requests to
# the url exposed by an APIRule.
#
# The following process is executed:
# 1. Start the zero downtime requests in the background. The requests will be sent once the APIRule is ready and the
# exposed host is reachable. The requests will be sent in a loop until the APIRule is deleted.
#  - Wait for 5 min until APIRule exists
#  - Wait for 5 min until APIRule status is OK
#  - If handler is jwt, wait for 1 min until the OAuth2 mock server is available and get the bearer token
#  - Wait for 1 min until the host in the APIRule is available
#  - Send requests in parallel to the exposed host until the requests fail and in this case check if the APIRule
#    still exists to determine if the test failed or succeeded.
# 2. Run the godog test that will migrate the APIRule from v1beta1 to v2alpha1 parallel to the requests.
# 3. Check if the zero downtime requests were successful.

set -eo pipefail

# The following trap is useful when breaking the script (ctrl+c), so it stops also background jobs
trap 'kill $(jobs -p)' INT

PARALLEL_REQUESTS=5

HANDLER="$1"

if [[ -z "$HANDLER" || ! "$HANDLER" =~ ^(jwt|noop|no_auth|allow|oauth2_introspection)$ ]]; then
  echo "zero-downtime: Handler not provided or invalid. Must be one of: jwt, noop, no_auth, allow, oauth2_introspection"
  exit 1
fi

echo "zero-downtime: Running zero downtime tests for handler '$HANDLER'"

# Function to run zero downtime requests to the exposed host of the APIRule
run_zero_downtime_requests() {
  local handler="$1"
  local bearer_token=""

  # Wait until the APIRule from the test is created and in status OK
  # At the time of writing this script kubectl wait does not support waiting for a resource that doesn't exist yet.
  # This is supported in an upcoming kubectl version.
  wait_for_api_rule_to_exist
  echo "zero-downtime: APIRule found, waiting for status OK"

  kubectl wait --for='jsonpath={.status.APIRuleStatus.code}=OK' --timeout=5m apirules -A -l test=v1beta1-migration
  echo "zero-downtime: APIRule status is OK"

  # Get the host set in the APIRule
  exposed_host=$(kubectl get apirules -A -l test=v1beta1-migration -o jsonpath='{.items[0].spec.host}')
  echo "zero-downtime: APIRule host is ${exposed_host}"

  local url_under_test="https://${exposed_host}/headers"

  if [ "$handler" == "jwt" ] || [ "$handler" == "oauth2_introspection" ]; then
    if [ -z "${TEST_OIDC_CONFIG_URL}" ]; then
      echo "zero-downtime: No OIDC_CONFIG_URL provided, assuming oauth mock"
      # Wait until the OAuth2 mock server host is available
      wait_for_url "https://oauth2-mock.${TEST_DOMAIN}/.well-known/openid-configuration"
      token_url="https://oauth2-mock.${TEST_DOMAIN}/oauth2/token"

      # Get the access token from the OAuth2 mock server
      echo "zero-downtime: Getting access token from URL '$token_url'"
      bearer_token=$(curl --fail --silent -kX POST "$token_url" -d "grant_type=client_credentials" -d "token_format=jwt" \
        -H "Content-Type: application/x-www-form-urlencoded" | jq -r ".access_token")
    else
      if [ -z "$TEST_CLIENT_ID" ] || [ -z "$TEST_CLIENT_SECRET" ]; then
        echo "No client ID or secret, failing"
        exit 3
      fi
      echo "zero-downtime: TEST_OIDC_CONFIG_URL provided, getting token url"
      token_url=$(curl --fail --silent "${TEST_OIDC_CONFIG_URL}" | jq -r .token_endpoint)
      if [ -z "$token_url" ]; then
        echo "Can't get token url"
        exit 4
      fi

      echo "zero-downtime: Getting access token"
      bearer_token=$(curl --fail --silent -kX POST "$token_url" -u "${TEST_CLIENT_ID}:${TEST_CLIENT_SECRET}" -d "grant_type=client_credentials" -d "token_format=jwt" \
        -H "Content-Type: application/x-www-form-urlencoded" | jq -r ".access_token")
    fi
  fi

  # Wait until the host in the APIRule is available. This may take a very long time because the httpbin application
  # used in the integration tests takes a very long time to start successfully processing requests, even though it is
  # already ready.
  echo "zero-downtime: Waiting until $url_under_test is available"
  wait_for_url "$url_under_test" "$bearer_token"

  echo "zero-downtime: Sending requests to ${url_under_test} in ${PARALLEL_REQUESTS} parallel threads"

  # Run the send_requests function in parallel processes
  for (( i = 0; i < PARALLEL_REQUESTS; i++ )); do
    echo "zero-downtime: Starting thread $i"
    send_requests "$i" "$url_under_test" "$bearer_token" &
    request_pids[$i]=$!
  done

  # Wait for all send_requests processes to finish or fail fast if one of them fails
  for pid in "${request_pids[@]}"; do
    wait "${pid}" && request_runner_exit_code=$? || request_runner_exit_code=$?
    if [ "${request_runner_exit_code}" -ne 0 ]; then
        echo "zero-downtime: A sending requests subprocess failed with a non-zero exit status."
        exit 1
    fi
  done

  exit 0
}

wait_for_api_rule_to_exist() {
  local attempts=1
  echo "zero-downtime: Waiting for the APIRule to exist"
  # Wait for 5min
  while [[ $attempts -le 300 ]] ; do
    apirule=$(kubectl get apirules -A -l test=v1beta1-migration --ignore-not-found) && kubectl_exit_code=$? || kubectl_exit_code=$?
    if [ "${kubectl_exit_code}" -ne 0 ]; then
        echo "zero-downtime: kubectl failed when listing apirules, exit code: $kubectl_exit_code"
        exit 2
    fi
  	[[ -n "$apirule" ]] && return 0
  	sleep 1
    ((attempts = attempts + 1))
  done
  echo "zero-downtime: APIRule not found"
  exit 1
}

wait_for_url() {
  local url="$1"
  local bearer_token="${2:-''}"
  local attempts=1

  echo "zero-downtime: Waiting for URL '${url}' to be available"

  # Wait for 10min
  while [[ $attempts -le 600 ]] ; do
    curl_exit_code=0
    curl_response=$(curl --include --silent  --fail --insecure --location -H "x-ext-authz: allow" -H "Authorization: Bearer ${bearer_token}" "${url}") || curl_exit_code=$?
    if [ "${curl_exit_code}" -eq 0 ]; then
      echo "zero-downtime: ${url} is available for requests"
  	  return 0
    fi
  	sleep 1
    ((attempts = attempts + 1))
  done

  echo "zero-downtime: ${url} is not available for requests, curl exit code: ${curl_exit_code}, response: "$'\n'"${curl_response}"
  exit 1
}

# Function to send requests to a given url with optional bearer token
send_requests() {
  local thread="$1"
  local url="$2"
  local bearer_token="$3"
  local request_count=0
  echo "zero-downtime: thread ${thread}, sending requests to ${url}"

  while true; do
    ((request_count = request_count + 1))

    curl_exit_code=0
    curl_response=""
    if [ -n "$bearer_token" ]; then
      curl_response=$(curl --include --silent  --fail --insecure -H "x-ext-authz: allow" -H "Authorization: Bearer $bearer_token" "$url") || curl_exit_code=$?
    else
      curl_response=$(curl --include --silent  --fail --insecure -H "x-ext-authz: allow" "$url") || curl_exit_code=$?
    fi

    if [ "${curl_exit_code}" -ne 0 ]; then
      echo "zero-downtime: thread ${thread}, request ${request_count}, curl exit code: ${curl_exit_code}, curl response:"$'\n'"${curl_response}"

      # If there is an error and the APIRule still exists, the test is failed, but if an error is received only when the
      # APIRule is deleted, the test is successful, because without an APIRule the request must fail as no host
      # is exposed. This was the most reliable way to detect when to stop the requests, since only sending requests
      # when the APIRule exists led to flaky results.

      # Let's sleep here few seconds to avoid the race condition - when the APIRule still exists, but is being deleted
      # (so it is not effective anymore)
      sleep 2
      if kubectl get apirules -A -l test=v1beta1-migration --ignore-not-found | grep -q .; then
        echo "zero-downtime: thread ${thread}, test failed after ${request_count} requests. Canceling requests because of curl exit code ${curl_exit_code}"
        exit 1
      elif [ $request_count -lt 10 ]; then
        echo "zero-downtime: thread ${thread}, there were less than 10 requests, something was wrong with the tests"
        exit 2
      else
        echo "zero-downtime: thread ${thread}, test successful after ${request_count} requests. Stopping requests because APIRule is deleted."
        exit 0
      fi
    fi
  done
}

start() {
  local handler="$1"

  # Start the requests in the background as soon as the APIRule is ready
  run_zero_downtime_requests "$handler" &
  zero_downtime_requests_pid=$!

  echo "zero-downtime: Starting integration test scenario for handler '$handler'"

  go test -count=1 -timeout 15m ./tests/integration -v -race -run "TestOryJwt/Migrate_v1beta1_APIRule_with_${handler}_handler" && test_exit_code=$? || test_exit_code=$?
  if [ "${test_exit_code}" -ne 0 ]; then
    echo "zero-downtime: Test execution failed"
    return 1
  fi

  wait $zero_downtime_requests_pid && zero_downtime_exit_code=$? || zero_downtime_exit_code=$?
  if [ "${zero_downtime_exit_code}" -ne 0 ]; then
    echo "zero-downtime: Requests returned a non-zero exit status, that means requests failed or returned a status not equal 200"
    return 2
  fi

  echo "zero-downtime: Test completed successfully"
  return 0
}

start "$HANDLER" && start_exit_code="$?" || start_exit_code="$?"
if [ "$start_exit_code" == "1" ]; then
  echo "zero-downtime: godog integration tests failed"
  exit 1
elif [ "$start_exit_code" == "2" ]; then
  echo "zero-downtime: Zero-downtime requests failed"
  exit 2
fi

echo "zero-downtime: Tests successful"
exit 0

#!/bin/bash

set -o nounset

PARALLEL_REQUESTS=5

# Script to run zero downtime tests by executing one godog integration test and sending requests to the endpoint
# exposed by an APIRule.
#
# The following process is executed:
# 1. Start the zero downtime requests in the background. The will be sent once the APIRule is ready and until the
#    APIRule is deleted.
# 2. Run the godog test that will migrate the APIRule from v1beta1 to v2alpha1.
# 3. Check if the zero downtime requests were successful.

run_zero_downtime_requests() {
  local handler="$1"
  local bearerToken=""

  # Wait until the APIRule created in the test is in status OK
  wait_for_api_rule_to_exist
  echo "zero-downtime: APIRule found"
  kubectl wait --for='jsonpath={.status.APIRuleStatus.code}=OK' --timeout=5m apirules -A -l test=v1beta1-migration
  exposed_host=$(kubectl get apirules -A -l test=v1beta1-migration -o jsonpath='{.items[0].spec.host}')
  url_under_test="https://$exposed_host/anything"

  if [ "$handler" == "jwt" ]; then
    echo "zero-downtime: Getting access token"
    # Get the access token from the OAuth2 mock server
    tokenUrl="https://oauth2-mock.$KYMA_DOMAIN/oauth2/token"
    # TODO get cluster domain and token from the OAuth2 mock server
  fi

  echo "zero-downtime: Waiting for the new host to be propagated"
  # Propagation of the new host can take some time, therefore there is a wait for 30 secs,
  # even though APIRule is in OK state.
  sleep 30
  echo "zero-downtime: Sending requests to $url_under_test"

  # Run the send_requests function in parallel child processes
  for (( i = 0; i < PARALLEL_REQUESTS; i++ )); do
    send_requests "$url_under_test" "$bearerToken" &
    request_pids[$i]=$!
  done

  for pid in ${request_pids[*]}; do
    wait $pid
    if [ $? -ne 0 ]; then
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
  while [[ $attempts -le 3000 ]] ; do
  	apirule=$(kubectl get apirules -A -l test=v1beta1-migration --ignore-not-found)
  	[[ -n "$apirule" ]] && return 0
  	sleep 0.1
    ((attempts = attempts + 1))
  done
  echo "zero-downtime: APIRule not found"
}


# TODO: Add handling of requests that must send a bearer token
# Function to send requests to a given endpoint
send_requests() {
  local endpoint="$1"
  local bearerToken="$2"

  # Stop sending requests when the APIRule is deleted to avoid false negatives by sending requests
  # to an endpoint that is no longer exposed.
  while kubectl get apirules -A -l test=v1beta1-migration --ignore-not-found | grep -q .; do

    if [ -n "$bearerToken" ]; then
      response=$(curl -fsSk -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $bearerToken" "$endpoint")
    else
      response=$(curl -fsSk -o /dev/null -w "%{http_code}" "$endpoint")
    fi

    if [ "$response" -ne 200 ]; then
      echo "zero-downtime: Error, received HTTP status code $response"
      exit 1
    fi
  done

  exit 0
}

start() {
  local handler="$1"

  # Start the requests in the background as soon as the APIRule is ready
  run_zero_downtime_requests "$handler" &
  zero_downtime_requests_pid=$!

  echo "zero-downtime: Starting integration test scenario for handler $handler"
  case $handler in
    "no_auth")
      make TEST=Migrate_v1beta1_APIRule_with_no_auth_handler test-migration-zero-downtime
      ;;
    "noop")
      make TEST=Migrate_v1beta1_APIRule_with_noop_handler test-migration-zero-downtime
      ;;
    "allow")
      make TEST=Migrate_v1beta1_APIRule_with_allow_handler test-migration-zero-downtime
      ;;
    "jwt")
      make TEST=Migrate_v1beta1_APIRule_with_jwt_handler test-migration-zero-downtime
      ;;
    *)
      echo "Invalid handler specified"
      exit 1
      ;;
  esac

  wait $zero_downtime_requests_pid
  if [ $? -ne 0 ]; then
    echo "zero-downtime: Requests returned a non-zero exit status, that means requests failed or returned a status not equal 200"
    return 1
  fi

  echo "zero-downtime: Test completed successfully"
  return 0
}

start "allow"
allow_exit_code=$?
start "noop"
noop_exit_code=$?

start "no_auth"
no_auth_exit_code=$?

#start "jwt
#jwt_exit_code=$?


echo "zero-downtime: allow exit code: $allow_exit_code"
echo "zero-downtime: noop exit code: $noop_exit_code"
echo "zero-downtime: no_auth exit code: $no_auth_exit_code"
#echo "zero-downtime: jwt_exit_code: $jwt_exit_code"

# TODO exit code handling

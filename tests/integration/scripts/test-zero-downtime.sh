#!/bin/bash

set -o nounset
set -o pipefail

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

  # Wait until the APIRule created in the test is in status OK
  wait_for_api_rule_to_exist
  echo "zero-downtime: APIRule found"
  kubectl wait --for='jsonpath={.status.APIRuleStatus.code}=OK' --timeout=5m apirules -A -l test=v1beta1-migration
  exposed_host=$(kubectl get apirules -A -l test=v1beta1-migration -o jsonpath='{.items[0].spec.host}')
  url_under_test="https://$exposed_host/anything"

  echo "zero-downtime: Waiting for the new host to be propagated"
  # Propagation of the new host can take some time, therefore there is a wait for 30 secs,
  # even though APIRule is in OK state.
  sleep 30
  echo "zero-downtime: Sending requests to $url_under_test"

  # Run the send_requests function in parallel child processes
  for (( i = 0; i < $PARALLEL_REQUESTS; i++ )); do
    send_requests $url_under_test &
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

  # Stop sending requests when the APIRule is deleted to avoid false negatives by sending requests
  # to an endpoint that is no longer exposed.
  while kubectl get apirules -A -l test=v1beta1-migration --ignore-not-found | grep -q .; do
    response=$(curl -fsSk -o /dev/null -w "%{http_code}" "$endpoint")
    if [ "$response" -ne 200 ]; then
      echo "zero-downtime: Error, received HTTP status code $response"
      exit 1
    fi
  done

  exit 0
}

# Function to run tests
run_test() {
  echo "zero-downtime: Starting integration test scenario"
  #make TEST=Migrate_v1beta1_APIRule_with_no_auth_handler test-migration-zero-downtime
  #make TEST=Migrate_v1beta1_APIRule_with_noop_handler test-migration-zero-downtime
  make TEST=Migrate_v1beta1_APIRule_with_allow_handler test-migration-zero-downtime
}

# Start the requests in the background as soon as the APIRule is ready
run_zero_downtime_requests &
zero_downtime_requests_pid=$!

run_test

wait $zero_downtime_requests_pid
if [ $? -ne 0 ]; then
  echo "zero-downtime: Requests returned a non-zero exit status, that means requests failed or returned a status not equal 200"
  exit 1
fi

echo "zero-downtime: Test completed successfully"

exit 0
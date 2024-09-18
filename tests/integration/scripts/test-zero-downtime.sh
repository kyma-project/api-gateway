#!/bin/bash

set -o nounset
set -o pipefail

if [ $# == 0 ]; then
    echo "Usage: $0 param1"
    echo "* param1: URL that will be tested"
    exit 1
fi

url_under_test="$1"

# TODO: Add handling of requests that must send a bearer token
# Function to send requests to a given endpoint
send_requests() {
  local endpoint="$1"
  while true; do
    response=$(curl -fsSk -o /dev/null -w "%{http_code}" "$endpoint")
    if [ "$response" -ne 200 ]; then
      echo "Error: Received HTTP status code $response"
      exit 1
    fi
  done
}

# Function to run tests
run_tests() {
  sleep 2  # Replace with actual test execution logic
}

echo "Sending requests to $url_under_test"
# Run the send_requests function in the background
send_requests $url_under_test &
request_pid=$!

# Run the tests
run_tests

# Attempt to stop the send_requests function
if ! kill $request_pid 2>/dev/null; then
  echo "Requests failed or returned a status not equal 200"
  exit 2
fi

echo "Test completed successfully"
#!/bin/bash

set -eo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'

if [ -z "$PULL_PULL_SHA" ]; then
  echo "Skipping code coverage guard execution, not on PR commit"
  exit 0
fi

if [ -z "$PULL_BASE_SHA" ]; then
  echo "Base commit for main not set as environment variable"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"
cd "${SCRIPT_DIR}/../../../" || exit 1

echo -e "Code coverage guard (ensures PRs do not lower code coverage)"
echo -e "Running tests on: PR-${PULL_NUMBER} (${PULL_PULL_SHA})"
make test

if [[ $? != 0 ]]; then
	echo -e "${RED}✗ make test\\n${NC}"
	exit 1
else
    coverage_pr=$(go tool cover -func=cover.out | grep total | awk '{print substr($3,1,length($3)-1)}')
	echo -e "Total coverage on PR-${PULL_NUMBER}: ${coverage_pr}%"
	echo -e "${GREEN}√ make test${NC}"
    rm cover.out
fi

git fetch && git checkout $PULL_BASE_SHA

echo -e "Running tests on: main (${PULL_BASE_SHA})"
make test

if [[ $? != 0 ]]; then
	echo -e "${RED}✗ make test\\n${NC}"
	exit 1
else
    coverage_main=$(go tool cover -func=cover.out | grep total | awk '{print substr($3,1,length($3)-1)}')
	echo -e "Total coverage on main: ${coverage_main}%"
	echo -e "${GREEN}√ make test${NC}"
    rm cover.out
fi

if awk "BEGIN {exit !($coverage_pr >= $coverage_main)}"; then
	echo -e "${GREEN}√ Thanks for keeping & increasing code coverage!"
else
	echo -e "${RED}✗ This PR lowering code coverage compared to main branch! Please add tests.\\n${NC}"
	exit 1
fi

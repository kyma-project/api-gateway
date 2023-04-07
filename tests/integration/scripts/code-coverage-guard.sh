#!/bin/bash

set -eo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

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

echo -e "Code Coverage Guard - Ensures PRs do not lower code coverage\\n"
echo -e "Running tests on: PR-${PULL_NUMBER} (${PULL_PULL_SHA})\\n"

if ! make test; then
	echo -e "${RED}✗ make test\\n${NC}"
	exit 1
else
    coverage_pr=$(go tool cover -func=cover.out | grep total | awk '{print substr($3,1,length($3)-1)}')
	echo -e "Total coverage on PR-${PULL_NUMBER}: ${coverage_pr}%\\n"
	echo -e "${GREEN}√ make test\\n${NC}"
    rm cover.out
fi

git checkout . && git clean -xffd
git fetch && git checkout $PULL_BASE_SHA

echo -e "Running tests on: main (${PULL_BASE_SHA})\\n"

if ! make test; then
	echo -e "${RED}✗ make test\\n${NC}"
	exit 1
else
    coverage_main=$(go tool cover -func=cover.out | grep total | awk '{print substr($3,1,length($3)-1)}')
	echo -e "Total coverage on main: ${coverage_main}%\\n"
	echo -e "${GREEN}√ make test\\n${NC}"
    rm cover.out
fi

if awk "BEGIN {exit !($coverage_pr >= $coverage_main)}"; then
	echo -e "${GREEN}√ Thanks for keeping & increasing code coverage!\\n${NC}"
else
	echo -e "${RED}✗ This PR is lowering code coverage compared to main branch! Please add tests.\\n${NC}"
	exit 1
fi

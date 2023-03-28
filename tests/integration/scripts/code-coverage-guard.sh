#!/bin/bash

set -eox pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'

if [ -z "$PULL_PULL_SHA" ]; then
  echo "WORKAROUND: skip jobguard execution - not on PR commit"
  exit 0
fi

cd "${KYMA_PROJECT_DIR}/api-gateway" || exit 1

echo -e "Code coverage guard (ensures PRs do not lower code coverage)"
echo -e "Running tests on PR: PR-${PULL_NUMBER}"
make test

if [[ $? != 0 ]]; then
	echo -e "${RED}✗ make test\\n${NC}"
	exit 1
else
    coverage_pr=$(go tool cover -func=cover.out | grep total | awk '{print $3}')
	echo -e "Total coverage on PR-${PULL_NUMBER}: ${coverage_pr}"
	echo -e "${GREEN}√ make test${NC}"
    rm cover.out
fi

git fetch
git reset --hard origin/main

echo -e "Running tests on main branch"
make test

if [[ $? != 0 ]]; then
	echo -e "${RED}✗ make test\\n${NC}"
	exit 1
else
    coverage_main=$(go tool cover -func=cover.out | grep total | awk '{print $3}')
	echo -e "Total coverage on main: ${coverage_main}"
	echo -e "${GREEN}√ make test${NC}"
    rm cover.out
fi

if (( $(echo "${coverage_pr} < ${coverage_main}" | bc -l) )); then
	echo -e "${RED}✗ This PR lowering code coverage compared to main branch! Please add tests.\\n${NC}"
	exit 1
else
	echo -e ""
	echo -e "${GREEN}√ Thanks for keeping / increasing code coverage!"
fi

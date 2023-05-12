#!/usr/bin/env bash
# meant to run in Prow env from the root of the repo

set -e

if [[ "$CI" != "true" ]]; then
    echo "Not in CI. Exiting..."
    exit 1
fi

GORELEASER_VERSION="${GORELEASER_VERSION:-v1.18.2}"

JOB_NAME_PATTERN="rel-.*-integration" tests/integration/scripts/jobguard.sh
curl -sfLo /tmp/goreleaser_Linux_x86_64.tar.gz "https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz"
tar xf /tmp/goreleaser_Linux_x86_64.tar.gz -C /tmp

git remote add origin "https://github.com/$REPO_OWNER/$REPO_NAME"
/tmp/goreleaser release

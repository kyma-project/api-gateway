#!/usr/bin/env bash

# Verifies PROJECT_ACCESS_CLASSIC by querying the GitHub GraphQL API
# for the project defined by PROJECT_NUMBER in the kyma-project organisation.

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars GH_TOKEN ORGANIZATION PROJECT_NUMBER

start_group "Query GitHub project ${ORGANIZATION}/${PROJECT_NUMBER}"
if ! gh api graphql -f query='
  query($org: String!, $number: Int!) {
    organization(login: $org) {
      projectV2(number: $number) {
        id
        title
      }
    }
  }' -f org="${ORGANIZATION}" -F number="${PROJECT_NUMBER}" > project_data.json; then
  >&2 echo "FAIL: GraphQL query failed — token may lack project read permissions"
  exit 1
fi

PROJECT_TITLE=$(jq -r '.data.organization.projectV2.title // empty' project_data.json)
if [ -z "${PROJECT_TITLE}" ]; then
  >&2 echo "FAIL: could not retrieve project title — token may lack access to project ${PROJECT_NUMBER}"
  cat project_data.json
  exit 1
fi
end_group

echo "PASS: PROJECT_ACCESS_CLASSIC is valid — project: \"${PROJECT_TITLE}\""

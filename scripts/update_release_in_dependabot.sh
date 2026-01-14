#!/usr/bin/env bash

# This script updates release branch in the dependabot yaml file

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

release_tag=$1
dependabot_file=$2

echo "Update release in dependabot: release tag: ${release_tag}"

major=$(echo "${release_tag}" | cut -d. -f1)
minor=$(echo "${release_tag}" | cut -d. -f2)
current_major_minor="${major}.${minor}"

version_regex='\(.*\"release-\)\(.*\)\(\".*\)'

major_minor_from_dependabot=$(sed -n "s|${version_regex}|\2|p" "${dependabot_file}" | head -n1)
if [ -z "${major_minor_from_dependabot}" ]; then
  echo "Can't get release version from dependabot file"
else
  expected_major_minor=$(printf '%s\n' "${major_minor_from_dependabot}" "${current_major_minor}" | sort -V | tail -n1)
  if [ "${expected_major_minor}" = "${major_minor_from_dependabot}" ]; then
    echo "Release version ${major_minor_from_dependabot} doesn't need to be updated"
  else
    echo "Release version ${major_minor_from_dependabot} needs to be updated to ${current_major_minor}"
    sed "s|${version_regex}|\1${current_major_minor}\3|g" "${dependabot_file}" > "${dependabot_file}".new
    mv "${dependabot_file}".new "${dependabot_file}"
  fi
fi

#!/usr/bin/env bash

# Description: This script returns the most recent commit ID (SHA) from the git repository (accessible via current dir)
# for which the Docker image <image>:<commit-sha> exists (can be pulled)
# It requires the parameters:
# 1 - fully qualified Docker image (without tag)

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=./common.sh
source "${script_dir}/common.sh"

require_positional image_name "$1"

found_image_commit_id=""
max_commits=10

for commit_id in $(git log -n "${max_commits}" --format=%H); do
    image_name_with_tag_to_check="${image_name}:${commit_id}"
    echo "Checking image ${image_name_with_tag_to_check}" >&2

    exit_code=0 && docker pull -q "${image_name_with_tag_to_check}" >&2 || exit_code=$?

    if [ "$exit_code" == "0" ]; then
        echo "Image ${image_name_with_tag_to_check} exists" >&2
        found_image_commit_id="${commit_id}"
        break
    else
        echo "Image ${image_name_with_tag_to_check} doesn't exist" >&2
    fi
done

if [ -z "${found_image_commit_id}" ]; then
    echo "No image found!" >&2
    exit 2
fi

echo "${found_image_commit_id}"

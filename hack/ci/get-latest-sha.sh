#!/usr/bin/env bash

# Description: This script returns the most recent commit ID (SHA) from the git repository (accessible via current dir)
# for which the Docker image <image>:<commit-sha> exists (can be pulled)
# It requires the parameters:
# 1 - fully qualified Docker image (without tag)

set -eo pipefail

found_image_commit_id=""
max_commits=10
image_name=$1

if [ -z "${image_name}" ]; then
    echo "Image name (without tag) must be provided as first parameter" >&2
    exit 1
fi

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

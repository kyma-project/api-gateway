#!/usr/bin/env bash

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly API_GATEWAY_DIR="/home/prow/go/src/github.com/kyma-project/api-gateway"

# locally delete the stable tag in order to have all the change logs listed under the new release version
delete_stable_tag() {
    if [[ $(git tag -l stable) ]]; then
        git tag -d stable
    fi
}

get_new_release_version() {
    # get the list of tags in a reverse chronological order
    TAG_LIST=($(git tag --sort=-creatordate))
    NEW_RELEASE_VERSION=${TAG_LIST[0]}
}

get_current_release_version() {
    # get the list of tags in a reverse chronological order excluding release candidate tags
    TAG_LIST_WITHOUT_RC=($(git tag --sort=-creatordate | grep -v -e "-rc"))
    if [[ $NEW_RELEASE_VERSION == *"-rc"* ]]; then
        CURRENT_RELEASE_VERSION=${TAG_LIST_WITHOUT_RC[0]}
    else
        CURRENT_RELEASE_VERSION=${TAG_LIST_WITHOUT_RC[1]}
    fi
}

main() {
    git remote add origin git@github.com:kyma-project/api-gateway.git
    delete_stable_tag
    get_new_release_version
    get_current_release_version
    # generate release changelog
    docker run --rm -v "${API_GATEWAY_DIR}":/repository -w /repository -e FROM_TAG="${CURRENT_RELEASE_VERSION}" -e NEW_RELEASE_TITLE="${NEW_RELEASE_VERSION}" -e GITHUB_AUTH="${BOT_GITHUB_TOKEN}" -e CONFIG_FILE=.github/package.json eu.gcr.io/kyma-project/changelog-generator:0.2.0 sh /app/generate-release-changelog.sh
    # release api-gateway with release notes generated by changelog-generator
    curl -sL https://git.io/goreleaser | VERSION=v0.118.2 bash  -s --  --release-notes .changelog/release-changelog.md
}

main

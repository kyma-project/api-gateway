#!/usr/bin/env bash

make manifests generate

if [[ $(git status --porcelain) ]]; then
  echo "There were changes to the CRDs and you did not run 'make manifests generate'"
  echo "Please run 'make manifests generate' or apply the generated diff and commit the changes."
  git --no-pager diff
  git --no-pager diff > git-diff.diff
  exit 1
fi
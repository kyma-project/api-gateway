#!/usr/bin/env bash

make manifests generate generate-crd-docs sync-vendors-crds

if [[ $(git status --porcelain) ]]; then
  echo "There were changes to generated files or vendored CRDs that were not committed."
  echo "Please run 'make manifests generate generate-crd-docs sync-vendors-crds' and commit the changes."
  git --no-pager diff
  git --no-pager diff > git-diff.diff
  exit 1
fi
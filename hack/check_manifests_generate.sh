#!/usr/bin/env bash

make manifests generate generate-apirule-crd

if [[ $(git status --porcelain) ]]; then
  echo "There were changes to the CRDs and you did not run 'make manifests generate generate-apirule-crd'."
  echo "Please run 'make manifests generate generate-apirule-crd' and commit the changes."
  git diff --exit-code
  git diff > git-diff.diff
  exit 1
fi

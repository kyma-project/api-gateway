#!/usr/bin/env bash

make manifests generate generate-crd-docs

if [[ $(git status --porcelain) ]]; then
  echo "There were changes to the CRDs and you did not regenerate the CRD file and documentation."
  echo "Please run 'make manifests generate generate-crd-docs' or apply the generated diff and commit the changes."
  git --no-pager diff
  git --no-pager diff > git-diff.diff
  exit 1
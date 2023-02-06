#!/usr/bin/env bash
set -e -o pipefail

docker pull restic/restic
RESTIC_VERSION=$(docker run --rm restic/restic version | cut -f 2 -d ' ')

function replace() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$1" "$2"
  else
    sed -i "$1" "$2"
  fi
}

# replace Docker
replace "s|restic/restic:.*|restic/restic:${RESTIC_VERSION}|g" Dockerfile

# replace Readme
replace "s|auto-restic:.*|auto-restic:${RESTIC_VERSION}|g" README.md

git tag "${RESTIC_VERSION}"
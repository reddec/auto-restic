#!/usr/bin/env bash
set -e

docker pull restic/restic
RESTIC_VERSION=$(docker run --rm restic/restic version | cut -f 2 -d ' ')

if [ "$(git describe --tags 2>/dev/null || true)" == "${RESTIC_VERSION}" ]; then
  echo "no updates"
  exit 1
fi

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
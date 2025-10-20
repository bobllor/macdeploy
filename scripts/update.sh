#/usr/bin/env bash

# Updates the project to the latest tagged release.

set -e

git fetch origin
git checkout $(git describe --tags $(git rev-list --tags --max-count=1))

bash build.sh

docker compose up -d
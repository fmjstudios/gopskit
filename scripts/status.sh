#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(git rev-parse --show-toplevel)
GO_VERSION=$(cat "$ROOT/go.mod" | grep -E "go 1." | awk '{ print $2 }')

echo GIT_COMMIT_SHA "$(git rev-parse --short=7 HEAD)"
echo GIT_BRANCH "$(git rev-parse --abbrev-ref HEAD)"
echo GO_VERSION "$GO_VERSION"

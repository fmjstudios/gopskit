#!/usr/bin/env bash
#
# shellcheck disable=SC2002

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(git rev-parse --show-toplevel)
GO_VERSION=$(cat "$ROOT/MODULE.bazel" | grep -Po '^go_sdk.download\(version = "\K\d+.\d+.\d*"\)$' | awk '{ print substr( $0, 1, length($0)-2 ) }')
PLATFORM=$(cat "/etc/os-release" | grep -Po "PRETTY_NAME=[\'\"](\K.[^\"\']*)[^\'\"]")
DATE=$(date --rfc-3339=seconds)

echo GIT_COMMIT_SHA "$(git rev-parse --short=7 HEAD)"
echo GIT_BRANCH "$(git rev-parse --abbrev-ref HEAD)"
echo GO_VERSION "$GO_VERSION"
echo BUILD_DATE "$DATE"
echo PLATFORM "$PLATFORM"

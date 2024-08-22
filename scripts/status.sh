#!/usr/bin/env bash

echo VERSION "$()"
echo GIT_COMMIT_SHA "$(git rev-parse --short=7 HEAD)"
echo GIT_BRANCH "$(git rev-parse --abbrev-ref HEAD)"
echo GO_VERSION "$(go version | awk '{ print $3 }')"
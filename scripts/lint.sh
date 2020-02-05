#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck source=scripts/config.sh
source "$(dirname "$0")/config.sh"

if ! "$GOBIN/golangci-lint" --version 2>/dev/null | grep -q "$GOLANGCILINT_VERSION"; then
    curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/v${GOLANGCILINT_VERSION}/install.sh" \
        | sh -s -- -b "$GOBIN" "v${GOLANGCILINT_VERSION}"
fi

"$GOBIN/golangci-lint" run --timeout 5m

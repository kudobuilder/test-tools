#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck source=scripts/config.sh
source "$(dirname "$0")/config.sh"

go generate ./...

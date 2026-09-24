#!/usr/bin/env bash

# Verifies that GARDENER_APIRULE_ACCESS is a valid base64-encoded inline PGP-signed message
# using the same gopenpgp/v3 library as the controller (supports OpenPGP v6 / RFC 9580).

set -eo pipefail
script_dir="$(dirname "$(readlink -f "$0")")"
# shellcheck source=../common.sh
source "${script_dir}/../common.sh"

require_vars GARDENER_APIRULE_ACCESS

start_group "Verify PGP signature"
go run "${script_dir}/verify-apirule-access/main.go"
end_group

echo "PASS: GARDENER_APIRULE_ACCESS signature is valid"

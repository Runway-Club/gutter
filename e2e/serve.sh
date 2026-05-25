#!/usr/bin/env bash
# Build the gutter CLI, then build + serve the e2e testapp on :8080.
# Playwright's webServer config launches this and waits for the URL.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/.." && pwd)"

cli="$(mktemp -d)/gutter"
( cd "$root" && go build -o "$cli" ./cmd/gutter )

cd "$here/testapp"
exec "$cli" run

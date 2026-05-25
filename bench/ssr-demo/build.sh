#!/usr/bin/env bash
# Build the SSR demo: one WASM client binary + host-rendered CSR/SSR pages.
set -euo pipefail
cd "$(dirname "$0")"

echo "== client WASM =="
GOOS=js GOARCH=wasm go build -o /tmp/ssrdemo.wasm .

echo "== render HTML (host) =="
go run ./ssrgen

WEXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
for d in dist-csr dist-ssr; do
  cp /tmp/ssrdemo.wasm "$d/app.wasm"
  cp "$WEXEC" "$d/wasm_exec.js"
done
echo "built dist-csr/ and dist-ssr/  (wasm: $(du -h /tmp/ssrdemo.wasm | cut -f1))"

#!/usr/bin/env bash
# Build the compute kernels to WASM with both Go std and TinyGo, into separate
# dist dirs (they ship incompatible wasm_exec.js runtimes so can't share a page).
set -euo pipefail
cd "$(dirname "$0")"

echo "== Go std =="
rm -rf dist-go && mkdir -p dist-go
GOOS=js GOARCH=wasm go build -o dist-go/app.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist-go/wasm_exec.js
cp index.html kernels.js dist-go/
echo "Go app.wasm: $(du -h dist-go/app.wasm | cut -f1)"

echo "== TinyGo =="
if command -v tinygo >/dev/null 2>&1; then
  rm -rf dist-tinygo && mkdir -p dist-tinygo
  tinygo build -o dist-tinygo/app.wasm -target wasm .
  cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" dist-tinygo/wasm_exec.js
  cp index.html kernels.js dist-tinygo/
  echo "TinyGo app.wasm: $(du -h dist-tinygo/app.wasm | cut -f1)"
else
  echo "tinygo not found — skipping"
fi

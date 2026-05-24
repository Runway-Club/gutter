# Gutter — Makefile for building and installing the CLI.
#
# Common targets:
#   make build     - build ./bin/gutter for the host platform
#   make install   - install the gutter binary to $(INSTALL_DIR)
#   make uninstall - remove the installed binary
#   make vet       - go vet for host + WASM targets
#   make wasm      - go build for GOOS=js GOARCH=wasm (sanity check)
#   make examples  - build the showcase example to WASM
#   make clean     - remove build artefacts
#   make tidy      - go mod tidy

GO         ?= go
BIN        ?= gutter
BIN_DIR    ?= bin
INSTALL_DIR ?= $(shell $(GO) env GOPATH)/bin
LDFLAGS    ?= -s -w

CLI_PKG := ./cmd/gutter

.PHONY: help build install uninstall vet wasm examples llms clean tidy run-counter

help:
	@echo "Gutter CLI Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make build       Build $(BIN_DIR)/$(BIN) for the host platform"
	@echo "  make install     Install $(BIN) to $(INSTALL_DIR)"
	@echo "  make uninstall   Remove the installed binary"
	@echo "  make vet         go vet host + WASM"
	@echo "  make wasm        GOOS=js GOARCH=wasm go build ./..."
	@echo "  make examples    Build examples/showcase to WASM"
	@echo "  make llms        Generate llms-full.md (single-file docs for AI agents)"
	@echo "  make tidy        go mod tidy"
	@echo "  make clean       Remove build artefacts"

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BIN_DIR)/$(BIN) $(CLI_PKG)
	@echo "✓ built $(BIN_DIR)/$(BIN)"

install: build
	@mkdir -p $(INSTALL_DIR)
	@install -m 0755 $(BIN_DIR)/$(BIN) $(INSTALL_DIR)/$(BIN)
	@echo "✓ installed to $(INSTALL_DIR)/$(BIN)"
	@echo "  (make sure $(INSTALL_DIR) is on your PATH)"

uninstall:
	@rm -f $(INSTALL_DIR)/$(BIN)
	@echo "✓ removed $(INSTALL_DIR)/$(BIN)"

vet:
	$(GO) vet ./...
	GOOS=js GOARCH=wasm $(GO) vet ./...

wasm:
	GOOS=js GOARCH=wasm $(GO) build ./...

examples:
	cd examples/showcase && GOOS=js GOARCH=wasm $(GO) build -o app.wasm .

llms:
	$(GO) run scripts/gen_llms.go

tidy:
	$(GO) mod tidy

clean:
	@rm -rf $(BIN_DIR) dist app.wasm
	@find examples -name 'app.wasm' -delete 2>/dev/null || true
	@echo "✓ cleaned"

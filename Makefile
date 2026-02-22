SHELL := /bin/sh

GO ?= go
GOCACHE_DIR ?= /tmp/karl-go-cache
KARL_BIN ?= ./karl
WASM_OUT ?= assets/playground/karl.wasm
VSCODE_EXT_DIR ?= karl-vscode
VSCODE_CLI ?= code
CURSOR_CLI ?= cursor
GO_CMD = GOCACHE=$(GOCACHE_DIR) $(GO)
LATEST_VSIX = $$(ls -t $(VSCODE_EXT_DIR)/*.vsix 2>/dev/null | head -1)

.PHONY: help build build-karl build-wasm build-all test test-nocache lint examples workflow ci clean \
	vscode-package vscode-install vscode-install-cursor vscode-reinstall

help:
	@echo "Karl dev commands:"
	@echo "  make build         # build karl binary (./karl)"
	@echo "  make build-wasm    # rebuild playground wasm ($(WASM_OUT))"
	@echo "  make build-all     # build binary + wasm"
	@echo "  make test          # run go tests"
	@echo "  make test-nocache  # run go tests with cache disabled"
	@echo "  make lint          # run golangci-lint"
	@echo "  make examples      # run examples runtime suite"
	@echo "  make workflow      # run workflow contrib suite"
	@echo "  make ci            # local CI sequence"
	@echo "  make vscode-package        # package VS Code extension (.vsix)"
	@echo "  make vscode-install        # reinstall extension in VS Code"
	@echo "  make vscode-install-cursor # reinstall extension in Cursor"
	@echo "  make vscode-reinstall      # reinstall extension in VS Code + Cursor"

build: build-karl

build-karl:
	$(GO_CMD) build -o karl .

build-wasm:
	GOCACHE=$(GOCACHE_DIR) GOOS=js GOARCH=wasm $(GO) build -o $(WASM_OUT) ./wasm

build-all: build-karl build-wasm

test:
	$(GO_CMD) test ./...

test-nocache:
	$(GO_CMD) clean -testcache
	$(GO_CMD) test -count=1 ./...

lint:
	golangci-lint run --timeout=5m --out-format=colored-line-number

examples: build-karl
	KARL_BIN=$(KARL_BIN) scripts/run_examples_runtime.sh

workflow:
	cd examples/contrib/workflow && ./run_all_tests.sh

ci: test-nocache lint examples workflow

clean:
	rm -f karl
	$(GO_CMD) clean -testcache

vscode-package:
	cd $(VSCODE_EXT_DIR) && npm install && npm run package

vscode-install: vscode-package
	@VSIX="$(LATEST_VSIX)"; \
	if [ -z "$$VSIX" ]; then echo "No .vsix found in $(VSCODE_EXT_DIR)"; exit 1; fi; \
	$(VSCODE_CLI) --install-extension "$$VSIX" --force

vscode-install-cursor: vscode-package
	@VSIX="$(LATEST_VSIX)"; \
	if [ -z "$$VSIX" ]; then echo "No .vsix found in $(VSCODE_EXT_DIR)"; exit 1; fi; \
	$(CURSOR_CLI) --install-extension "$$VSIX" --force

vscode-reinstall: vscode-install vscode-install-cursor

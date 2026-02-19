SHELL := /bin/sh

GO ?= go
GOCACHE_DIR ?= /tmp/karl-go-cache
KARL_BIN ?= ./karl
WASM_OUT ?= assets/playground/karl.wasm
GO_CMD = GOCACHE=$(GOCACHE_DIR) $(GO)

.PHONY: help build build-karl build-wasm build-all test test-nocache lint examples workflow ci clean

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

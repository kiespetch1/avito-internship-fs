GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: build run test lint fmt tidy clean generate

build:
	$(GO) build -o bin/api ./cmd/api

generate:
	$(GO) tool oapi-codegen --config=internal/api/cfg.yaml ../api.yaml

run:
	$(GO) run ./cmd/api

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		echo "golangci-lint is not installed; ran go vet only"; \
	fi

fmt:
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) fmt ./...; \
	else \
		gofmt -w .; \
	fi

tidy:
	$(GO) mod tidy

clean:
	rm -rf bin coverage.out

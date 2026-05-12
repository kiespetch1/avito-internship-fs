GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: build run test test-e2e lint fmt tidy clean generate

build:
	$(GO) build -o bin/api ./cmd/api

generate:
	$(GO) tool oapi-codegen --config=internal/api/cfg.yaml ../api.yaml

run:
	$(GO) run ./cmd/api

test:
	$(GO) test ./...

# Requires a running Docker daemon (testcontainers spins up Postgres).
test-e2e:
	$(GO) test -tags=e2e -count=1 -timeout=180s ./cmd/api/...

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

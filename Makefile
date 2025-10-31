APP := goembedx
PKG := ./...
EXAMPLE := ./examples/basic.go
CLI := ./cmd/goembedx
COVER_FILE := coverage.out

.PHONY: all fmt lint test bench cover build example clean

all: fmt lint test

## ---------- Dev Commands ----------
fmt:
	@echo "🧹 Formatting code..."
	go fmt $(PKG)

lint:
	@echo "🔍 Running basic lint (go vet)..."
	go vet $(PKG)

test:
	@echo "✅ Running tests with race detector..."
	go test ./... -race -coverprofile=$(COVER_FILE) -covermode=atomic

bench:
	@echo "🏎️ Benchmarking vector ops..."
	go test -bench=. -benchmem ./vector

cover: test
	@echo "📊 Coverage report at $(COVER_FILE)"
	go tool cover -html=$(COVER_FILE)

## ---------- Build ----------
build:
	@echo "🔧 Building CLI..."
	go build -o bin/$(APP) $(CLI)

example:
	@echo "▶️ Running example..."
	go run $(EXAMPLE)

## ---------- Utilities ----------
clean:
	@echo "🧽 Cleaning workspace..."
	rm -rf bin/
	rm -f $(COVER_FILE)

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  fmt       Format code"
	@echo "  lint      Static analysis"
	@echo "  test      Tests w/ race + coverage"
	@echo "  bench     Run benchmarks"
	@echo "  cover     Open coverage UI"
	@echo "  build     Build CLI"
	@echo "  example   Run example program"
	@echo "  clean     Clean build artifacts"

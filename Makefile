.PHONY: build build-release clean run lint test deps

BINARY = llmtop
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"
CMD = ./cmd/llm-top

build:
	CGO_ENABLED=1 go build $(LDFLAGS) -trimpath -o $(BINARY) $(CMD)

bench:
	go test -bench=. -benchmem -count=3 ./internal/ui/...

pgo:
	@echo "Generating PGO profile..."
	go test -bench=. -benchmem -count=5 -cpuprofile=/tmp/pgo.pprof ./internal/ui/
	mv /tmp/pgo.pprof default.pgo
	@echo "PGO profile generated: default.pgo ($(shell ls -l default.pgo | awk '{print $$5}') bytes)"

build-release:
	CGO_ENABLED=0 go build $(LDFLAGS) -trimpath -o $(BINARY) $(CMD)
	strip $(BINARY) 2>/dev/null || true
	upx --lzma $(BINARY) 2>/dev/null || echo "upx not installed, skipping compression"

clean:
	rm -f $(BINARY)
	go clean

run: build
	./$(BINARY)

lint:
	go vet ./...

test:
	go test ./... -v

deps:
	go mod tidy

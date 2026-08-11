.PHONY: build test lint vendor-specs install clean

BINARY := kalshi-perp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/kalshi-perp

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/kalshi-perp

test:
	go test ./...

lint:
	go vet ./...

vendor-specs:
	curl -fsSL https://docs.kalshi.com/perps_openapi.yaml -o docs/openapi/perps_openapi.yaml
	curl -fsSL https://docs.kalshi.com/perps_asyncapi.yaml -o docs/asyncapi/perps_asyncapi.yaml

clean:
	rm -f $(BINARY)

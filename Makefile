.PHONY: build test lint vendor-specs install clean release-snapshot dist

BINARY := kalshi-perp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST := dist

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/kalshi-perp

install:
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/kalshi-perp

test:
	go test ./...

lint:
	go vet ./...

# Cross-compile release matrix without publishing (no goreleaser required).
dist:
	@mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_linux_amd64  ./cmd/kalshi-perp
	CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_linux_arm64  ./cmd/kalshi-perp
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_darwin_arm64 ./cmd/kalshi-perp
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)_darwin_amd64 ./cmd/kalshi-perp
	@ls -la $(DIST)

# Local GoReleaser snapshot (requires goreleaser installed).
release-snapshot:
	goreleaser release --snapshot --clean

vendor-specs:
	curl -fsSL https://docs.kalshi.com/perps_openapi.yaml -o docs/openapi/perps_openapi.yaml
	curl -fsSL https://docs.kalshi.com/perps_asyncapi.yaml -o docs/asyncapi/perps_asyncapi.yaml

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)


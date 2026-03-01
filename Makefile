PKG      := ./cmd/scan
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w

# Auto-detect host OS so the default build target always uses explicit GOOS/GOARCH.
# This prevents a GOOS=linux environment variable from silently producing a Linux
# ELF binary named scan.exe on a Windows machine.
ifeq ($(OS),Windows_NT)
  HOST_GOOS   := windows
  HOST_GOARCH := amd64
  HOST_BINARY := dist/scan.exe
else
  UNAME_S := $(shell uname -s)
  ifeq ($(UNAME_S),Darwin)
    HOST_GOOS   := darwin
    HOST_GOARCH := $(shell uname -m | sed 's/x86_64/amd64/')
    HOST_BINARY := dist/scan-darwin-$(HOST_GOARCH)
  else
    HOST_GOOS   := linux
    HOST_GOARCH := amd64
    HOST_BINARY := dist/scan-linux-amd64
  endif
endif

# Default: build for the detected host OS/arch with explicit flags -> dist/
.PHONY: build
build:
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) go build -ldflags "$(LDFLAGS)" -o $(HOST_BINARY) $(PKG)
	@echo "Built $(HOST_BINARY) (GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH))"

# Linux
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-linux-amd64 $(PKG)

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scan-linux-arm64 $(PKG)

# macOS
.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-darwin-amd64 $(PKG)

.PHONY: build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scan-darwin-arm64 $(PKG)

# Windows
.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-windows-amd64.exe $(PKG)

# All platforms
.PHONY: release
release: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows
	@echo ""
	@echo "Release binaries:"
	@ls -lh dist/scan-*

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -f dist/scan-*

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build               Build for detected host OS/arch (explicit GOOS/GOARCH) -> dist/"
	@echo "  build-linux         linux/amd64                -> dist/scan-linux-amd64"
	@echo "  build-linux-arm64   linux/arm64                -> dist/scan-linux-arm64"
	@echo "  build-darwin        darwin/amd64               -> dist/scan-darwin-amd64"
	@echo "  build-darwin-arm64  darwin/arm64 (Apple M*)    -> dist/scan-darwin-arm64"
	@echo "  build-windows       windows/amd64              -> dist/scan-windows-amd64.exe"
	@echo "  release             All platforms"
	@echo "  test                go test ./..."
	@echo "  vet                 go vet ./..."
	@echo "  clean               Remove build artifacts"

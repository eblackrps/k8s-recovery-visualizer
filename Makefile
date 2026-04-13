PKG           := ./cmd/scan
VERSION       := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE    ?= $(shell git show -s --format=%cs HEAD 2>/dev/null || echo "unknown")
MODEL_PKG     := k8s-recovery-visualizer/internal/model
LDFLAGS       := -s -w -X '$(MODEL_PKG).Version=$(VERSION)' -X '$(MODEL_PKG).BuildDate=$(BUILD_DATE)'
GUI_DIR       := ./desktop
FRONTEND_DIR  := $(GUI_DIR)/frontend
GUI_OUTPUT_BASE := K8V
NPM           ?= npm
PLAYWRIGHT    ?= npx playwright
WAILS         ?= wails
GOFMT_FILES   := $(shell git ls-files '*.go' ':!:desktop/frontend/*' ':!:vendor/*')

ifeq ($(OS),Windows_NT)
  HOST_GOOS   := windows
  HOST_GOARCH := amd64
  HOST_BINARY := dist/scan.exe
  GUI_OUTPUT  := $(GUI_OUTPUT_BASE).exe
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
  GUI_OUTPUT := $(GUI_OUTPUT_BASE)
endif

ifeq ($(HOST_GOOS),windows)
  GUI_PACKAGE_FLAGS ?= -nsis
  GUI_BUILD_FLAGS ?= -clean -skipbindings -s
else
  GUI_PACKAGE_FLAGS ?=
  GUI_BUILD_FLAGS ?= -clean -nopackage -skipbindings -s
endif

.PHONY: build build-cli build-gui package-gui dev-gui frontend-install frontend-build frontend-test screenshots
.PHONY: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows release-cli release
.PHONY: test race schema-validate schema-samples docker-build vet fmt smoke docs-check ci clean help

build: build-cli

build-cli:
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) go build -ldflags "$(LDFLAGS)" -o $(HOST_BINARY) $(PKG)
	@echo "Built $(HOST_BINARY) (GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH))"

frontend-install:
	cd $(FRONTEND_DIR) && $(NPM) ci

frontend-build:
	cd $(FRONTEND_DIR) && $(NPM) run build

frontend-test:
	cd $(FRONTEND_DIR) && $(NPM) test

screenshots: frontend-build
	cd $(FRONTEND_DIR) && $(PLAYWRIGHT) install chromium
	cd $(FRONTEND_DIR) && $(NPM) run screenshots

dev-gui:
	cd $(GUI_DIR) && $(WAILS) dev

build-gui: frontend-build
	cd $(GUI_DIR) && $(WAILS) build $(GUI_BUILD_FLAGS) -o $(GUI_OUTPUT) -ldflags "$(LDFLAGS)"

package-gui: frontend-build
	cd $(GUI_DIR) && $(WAILS) build -clean -skipbindings -s $(GUI_PACKAGE_FLAGS) -o $(GUI_OUTPUT) -ldflags "$(LDFLAGS)"

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-linux-amd64 $(PKG)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scan-linux-arm64 $(PKG)

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-darwin-amd64 $(PKG)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scan-darwin-arm64 $(PKG)

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-windows-amd64.exe $(PKG)

release-cli: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows
	@echo ""
	@echo "Release binaries written to dist/"

release: release-cli package-gui

fmt:
	gofmt -w $(GOFMT_FILES)

test:
	go test ./...

race:
	go test -race ./...

schema-validate:
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.0.0.schema.json -input ./out/recovery-scan.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/recovery-enriched.json

schema-samples:
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.0.0.schema.json -input ./schemas/examples/recovery-scan-3.0.0.sample.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.0.0.schema.json -input ./schemas/examples/recovery-scan-3.0.0.unverified.sample.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./schemas/examples/recovery-enriched-1.1.0.sample.json

smoke: build-cli
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "if (Test-Path 'out\\smoke') { Remove-Item -LiteralPath 'out\\smoke' -Recurse -Force }; New-Item -ItemType Directory -Force -Path 'out\\smoke' | Out-Null; & '.\\$(HOST_BINARY)' --dry-run --summary --runbook --redact --csv --out .\\out\\smoke --min-score 0"
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.0.0.schema.json -input ./out/smoke/recovery-scan.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/smoke/recovery-enriched.json
else
	rm -rf out/smoke
	mkdir -p out/smoke
	./$(HOST_BINARY) --dry-run --summary --runbook --redact --csv --out ./out/smoke --min-score 0
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.0.0.schema.json -input ./out/smoke/recovery-scan.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.1.0.schema.json -input ./out/smoke/recovery-enriched.json
endif

docker-build:
	docker build -t k8vis .

vet:
	go vet ./...

docs-check:
	go run ./tools/docscheck

ci: frontend-install frontend-build vet test race frontend-test smoke schema-samples docs-check build-gui

clean:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "foreach ($$path in @('dist','out\\smoke','desktop\\build\\bin','desktop\\frontend\\dist')) { if (Test-Path $$path) { Remove-Item -LiteralPath $$path -Recurse -Force } }"
else
	rm -rf dist/scan-* out/smoke desktop/build/bin desktop/frontend/dist/assets desktop/frontend/dist/index.html
endif

help:
	@echo "Targets:"
	@echo "  build               Build the CLI for the detected host OS/arch"
	@echo "  build-gui           Build the desktop app for the current host"
	@echo "  package-gui         Package the desktop app for the current host"
	@echo "  dev-gui             Run the Wails desktop app in dev mode"
	@echo "  frontend-install    Install frontend dependencies with npm ci"
	@echo "  frontend-build      Build the desktop frontend"
	@echo "  frontend-test       Run frontend tests"
	@echo "  screenshots         Generate deterministic GUI screenshots"
	@echo "  release-cli         Build all CLI release binaries"
	@echo "  release             Build CLI release binaries plus the current-host GUI package"
	@echo "  fmt                 Run gofmt -w across Go sources"
	@echo "  test                Run go test ./..."
	@echo "  race                Run go test -race ./..."
	@echo "  vet                 Run go vet ./..."
	@echo "  smoke               Run the dry-run smoke flow and schema validation"
	@echo "  schema-validate     Validate ./out artifacts against published schemas"
	@echo "  schema-samples      Validate committed schema examples"
	@echo "  docs-check          Validate local Markdown links and screenshot references"
	@echo "  docker-build        Build the container image locally"
	@echo "  ci                  Run the local validation stack"
	@echo "  clean               Remove generated build artifacts"

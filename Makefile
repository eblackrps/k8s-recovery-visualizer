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
NPM_PREFIX    := $(NPM) --prefix $(FRONTEND_DIR)
NPX_PREFIX    := npx --prefix $(FRONTEND_DIR)

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
  GUI_WINDOWS_TAGS ?= -tags native_webview2loader
else
  GUI_PACKAGE_FLAGS ?=
  GUI_BUILD_FLAGS ?= -clean -nopackage -skipbindings -s
  GUI_WINDOWS_TAGS ?=
endif

.PHONY: build build-cli build-gui package-gui dev-gui frontend-install frontend-build frontend-test screenshots
.PHONY: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-cli-cross release
.PHONY: test race schema-validate schema-samples docker-build vet fmt smoke docs-check ci clean help

build: build-cli

build-cli:
ifeq ($(OS),Windows_NT)
	go build -ldflags "$(LDFLAGS)" -o $(HOST_BINARY) $(PKG)
else
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) go build -ldflags "$(LDFLAGS)" -o $(HOST_BINARY) $(PKG)
endif
	@echo "Built $(HOST_BINARY) (GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH))"

frontend-install:
	$(NPM_PREFIX) ci

frontend-build:
	$(NPM_PREFIX) run build

frontend-test:
	$(NPM_PREFIX) test

screenshots: frontend-build
	$(NPX_PREFIX) playwright install chromium
	$(NPM_PREFIX) run screenshots

dev-gui:
	cd $(GUI_DIR) && $(WAILS) dev

build-gui: frontend-build
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "Set-Location 'desktop'; $(WAILS) build $(GUI_BUILD_FLAGS) $(GUI_WINDOWS_TAGS) -o $(GUI_OUTPUT) -ldflags \"$(LDFLAGS)\""
else
	cd $(GUI_DIR) && $(WAILS) build $(GUI_BUILD_FLAGS) -o $(GUI_OUTPUT) -ldflags "$(LDFLAGS)"
endif

package-gui: frontend-build
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$nsisPaths=@('C:\\Program Files (x86)\\NSIS','C:\\Program Files (x86)\\NSIS\\Bin'); $$existing=$$nsisPaths | Where-Object { Test-Path $$_ }; if ($$existing) { $$env:PATH=(($$existing -join ';') + ';' + $$env:PATH) }; Set-Location 'desktop'; $(WAILS) build -clean -skipbindings -s $(GUI_PACKAGE_FLAGS) $(GUI_WINDOWS_TAGS) -o $(GUI_OUTPUT) -ldflags \"$(LDFLAGS)\""
else
	cd $(GUI_DIR) && $(WAILS) build -clean -skipbindings -s $(GUI_PACKAGE_FLAGS) -o $(GUI_OUTPUT) -ldflags "$(LDFLAGS)"
endif

build-linux:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOOS='linux'; $$env:GOARCH='amd64'; go build -ldflags \"$(LDFLAGS)\" -o dist/scan-linux-amd64 $(PKG)"
else
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-linux-amd64 $(PKG)
endif

build-linux-arm64:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOOS='linux'; $$env:GOARCH='arm64'; go build -ldflags \"$(LDFLAGS)\" -o dist/scan-linux-arm64 $(PKG)"
else
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scan-linux-arm64 $(PKG)
endif

build-darwin:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOOS='darwin'; $$env:GOARCH='amd64'; go build -ldflags \"$(LDFLAGS)\" -o dist/scan-darwin-amd64 $(PKG)"
else
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-darwin-amd64 $(PKG)
endif

build-darwin-arm64:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$env:GOOS='darwin'; $$env:GOARCH='arm64'; go build -ldflags \"$(LDFLAGS)\" -o dist/scan-darwin-arm64 $(PKG)"
else
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scan-darwin-arm64 $(PKG)
endif

build-windows:
ifeq ($(OS),Windows_NT)
	go build -ldflags "$(LDFLAGS)" -o dist/scan-windows-amd64.exe $(PKG)
else
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scan-windows-amd64.exe $(PKG)
endif

build-cli-cross: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows
	@echo ""
	@echo "Cross-platform CLI binaries written to dist/"

release: package-gui
	@echo ""
	@echo "Current-host desktop package written by Wails. Public release artifacts are published by GitHub Actions."

fmt:
	gofmt -w $(GOFMT_FILES)

test:
	go test ./...

race:
	go test -race ./...

schema-validate:
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./out/recovery-scan.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.2.0.schema.json -input ./out/recovery-enriched.json

schema-samples:
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./schemas/examples/recovery-scan-3.1.0.sample.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./schemas/examples/recovery-scan-3.1.0.unverified.sample.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.2.0.schema.json -input ./schemas/examples/recovery-enriched-1.2.0.sample.json

smoke: build-cli
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "if (Test-Path 'out\\smoke') { Remove-Item -LiteralPath 'out\\smoke' -Recurse -Force }; New-Item -ItemType Directory -Force -Path 'out\\smoke' | Out-Null; & '.\\$(HOST_BINARY)' --dry-run --summary --runbook --redact --csv --out .\\out\\smoke --min-score 0"
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./out/smoke/recovery-scan.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.2.0.schema.json -input ./out/smoke/recovery-enriched.json
else
	rm -rf out/smoke
	mkdir -p out/smoke
	./$(HOST_BINARY) --dry-run --summary --runbook --redact --csv --out ./out/smoke --min-score 0
	go run ./cmd/schema-validate -schema ./schemas/recovery-scan-3.1.0.schema.json -input ./out/smoke/recovery-scan.json
	go run ./cmd/schema-validate -schema ./schemas/recovery-enriched-1.2.0.schema.json -input ./out/smoke/recovery-enriched.json
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
	@echo "  build-cli-cross     Build cross-platform CLI source binaries into dist/"
	@echo "  frontend-install    Install frontend dependencies with npm ci"
	@echo "  frontend-build      Build the desktop frontend"
	@echo "  frontend-test       Run frontend tests"
	@echo "  screenshots         Generate deterministic GUI screenshots"
	@echo "  release             Package the current-host desktop app"
	@echo "  fmt                 Run gofmt -w across Go sources"
	@echo "  test                Run go test ./..."
	@echo "  race                Run go test -race ./..."
	@echo "  vet                 Run go vet ./..."
	@echo "  smoke               Run the dry-run smoke flow and schema validation"
	@echo "  schema-validate     Validate ./out artifacts against published schemas"
	@echo "  schema-samples      Validate committed schema examples"
	@echo "  docs-check          Validate local Markdown links and screenshot references"
	@echo "  docker-build        Build the legacy local CLI container image (not a published release artifact)"
	@echo "  ci                  Run the local validation stack"
	@echo "  clean               Remove generated build artifacts"

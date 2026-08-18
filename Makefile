export CGO_ENABLED=0
export GOLANGCI_LINT_VERSION=2.12.2
export GORELEASER_VERSION=2.17.1

.PHONY: lint
lint:
	@echo "==> Running lints..."
	@golangci-lint run

.PHONY: fmt
fmt:
	@echo "==> Formatting code..."
	@golangci-lint fmt

.PHONY: unit
unit:
	@echo "==> Running tests..."
	@go test -v -cover ./...

.PHONY: test
test: lint unit

.PHONY: build
build:
	@echo "==> Building..."
	@go build -v ./...

.PHONY: release-check
release-check:
	@echo "==> Validating .goreleaser.yml..."
	@goreleaser check

.PHONY: snapshot
snapshot:
	@echo "==> Building a snapshot release..."
	@goreleaser build --snapshot --clean

.PHONY: upgrade-deps
upgrade-deps:
	@echo "==> Upgrading dependencies..."
	@go get -t -u ./...
	@go mod tidy

.PHONY: dependencies
dependencies:
	@echo "==> Downloading dependencies..."
	@go mod download

.PHONY: install
install: dependencies
	@echo "==> Building and installing odfe-alerts-handler binary..."
	@go install

.PHONY: tools
tools: tools.golangci-lint tools.goreleaser

# it's recommended to use the pre-built binary,
# see here - https://golangci-lint.run/usage/install/#other-ci
.PHONY: tools.golangci-lint
tools.golangci-lint:
	@echo "==> Installing Golangci-lint (v${GOLANGCI_LINT_VERSION}) ..."
	@wget -qO- -nv https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin v${GOLANGCI_LINT_VERSION}

# no pre-built install script upstream, and the release archives are per
# platform, so build from source to keep this portable
.PHONY: tools.goreleaser
tools.goreleaser:
	@echo "==> Installing GoReleaser (v${GORELEASER_VERSION}) ..."
	@go install github.com/goreleaser/goreleaser/v2@v${GORELEASER_VERSION}

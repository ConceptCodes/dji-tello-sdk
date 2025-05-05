# Location of the Go toolchain
GO              ?= go
BIN             ?= telloctl

# Version metadata for reproducible builds
VERSION         ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT      ?= $(shell git rev-parse --short HEAD)
BUILD_DATE      ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS         := -s -w \
                   -X "main.version=$(VERSION)" \
                   -X "main.commit=$(GIT_COMMIT)" \
                   -X "main.date=$(BUILD_DATE)"

############################
# House‑keeping targets
############################
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "%-15s %s\n", $$1, $$2}'

fmt:
	@$(GO) run mvdan.cc/gofumpt@latest -l -w .
	@$(GO) run golang.org/x/tools/cmd/goimports@latest -w .

tidy:
	$(GO) mod tidy

############################
# Build & install
############################
build: fmt tidy 
	$(GO) build -o bin/$(BIN) -ldflags '$(LDFLAGS)' ./cmd/telloctl

install: 
	$(GO) install ./cmd/telloctl

############################
# Testing & coverage
############################
test: 
	$(GO) test ./... -count=1

race: 
	$(GO) test -race ./...

cover: 
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

fuzz: 
	$(GO) test ./internal/parser -run=Fuzz -fuzztime=60s

############################
# Static analysis
############################
lint: 
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not installed. Run 'make install-lint' first." ; exit 1; }
	golangci-lint run ./...

install-lint: 
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $(shell go env GOPATH)/bin v1.56.0

vet:
	$(GO) vet ./...

staticcheck: ## staticcheck (extra analysis)
	@$(GO) run honnef.co/go/tools/cmd/staticcheck@latest ./...

############################
# CI convenience wrapper
############################
ci: lint vet staticcheck race cover ## Run all CI checks
	@echo "All checks passed!"

############################
# Release helpers
############################
tag: ## Tag a release (pass VERSION=x.y.z) e.g. make tag VERSION=v0.3.0
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

release: tag build ## Tag and build a release binary

clean: ## Remove built assets
	@rm -rf ./bin coverage.out

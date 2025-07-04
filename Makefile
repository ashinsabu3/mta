CURRENT_DIR := $(shell pwd)
DIST_DIR := $(CURRENT_DIR)/dist
CLI_NAME := mta
BIN_NAME := mta
CGO_FLAG := 0

HOST_OS := $(shell go env GOOS)
HOST_ARCH := $(shell go env GOARCH)

TARGET_ARCH ?= linux/amd64
TARGET_OS := $(shell echo $(TARGET_ARCH) | cut -d'/' -f1)
TARGET_ARCH_SHORT := $(shell echo $(TARGET_ARCH) | cut -d'/' -f2)

VERSION := $(shell cat ${CURRENT_DIR}/VERSION)

GOPATH ?= $(shell if test -x `which go`; then go env GOPATH; else echo "$(HOME)/go"; fi)
GOCACHE ?= $(HOME)/.cache/go-build
ARGOCD_LINT_GOGC ?= 20

default: build

$(DIST_DIR):
	mkdir -p $(DIST_DIR)

.PHONY: cli
cli-local:
	GOOS=$(HOST_OS) GOARCH=$(HOST_ARCH) CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME) .

.PHONY: build
build: $(DIST_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH_SHORT) CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME) .

.PHONY: run
run:
	go run main.go

.PHONY: lint
lint:
	golangci-lint --version
	# NOTE: If you get an "Out of Memory" error, try reducing GOGC value
	GOGC=$(ARGOCD_LINT_GOGC) GOMAXPROCS=2 golangci-lint run --fix --verbose

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: deps
deps:
	go list -m -u all

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)

.PHONY: release
release:
	@echo "Building release version $(VERSION)"
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH_SHORT) CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME)-$(VERSION) .

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build      Compile the project"
	@echo "  cli-local  Build for local system"
	@echo "  run        Run the Go application"
	@echo "  lint       Run golangci-lint"
	@echo "  test       Run unit tests"
	@echo "  fmt        Format the code"
	@echo "  tidy       Clean up go.mod dependencies"
	@echo "  deps       Check for outdated dependencies"
	@echo "  clean      Remove build files"
	@echo "  release    Build a release version"
	@echo "  help       Show this help message"

.PHONY: release-all
release-all: $(DIST_DIR)
	@echo "Building release version $(VERSION) for linux/amd64"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME)-$(VERSION)-linux-amd64 .

	@echo "Building release version $(VERSION) for linux/arm64"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME)-$(VERSION)-linux-arm64 .

	@echo "Building release version $(VERSION) for darwin/amd64"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME)-$(VERSION)-darwin-amd64 .

	@echo "Building release version $(VERSION) for darwin/arm64"
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_FLAG) go build -o $(DIST_DIR)/$(CLI_NAME)-$(VERSION)-darwin-arm64 .


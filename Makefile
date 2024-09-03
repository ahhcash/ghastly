# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Detect the operating system and architecture
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

# Binary name with OS and ARCH
BINARY_NAME=bin/vexdb

# Main package path
MAIN_PACKAGE=.

# Build flags
BUILD_FLAGS=-v

# Test flags
TEST_FLAGS=-v

.PHONY: all build clean test coverage deps vet fmt lint run help

all: build

build:
	GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

clean:
	$(GOCLEAN)
	rm -f vexdb-*

test:
	$(GOTEST) $(TEST_FLAGS) ./...

coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

deps:
	$(GOGET) ./...
	$(GOMOD) tidy

vet:
	$(GOCMD) vet ./...

fmt:
	$(GOCMD) fmt ./...

lint:
	golangci-lint run

run: build
	./$(BINARY_NAME)

build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o vexdb-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o vexdb-linux-arm64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o vexdb-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o vexdb-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o vexdb-windows-amd64.exe $(MAIN_PACKAGE)

help:
	@echo "Available targets:"
	@echo "  build     : Build the binary for the current OS and architecture"
	@echo "  clean     : Clean build artifacts"
	@echo "  test      : Run tests"
	@echo "  coverage  : Run tests with coverage"
	@echo "  deps      : Get dependencies and tidy go.mod"
	@echo "  vet       : Run go vet"
	@echo "  fmt       : Run go fmt"
	@echo "  lint      : Run golangci-lint"
	@echo "  run       : Build and run the binary"
	@echo "  build-all : Build binaries for multiple OS and architectures"
	@echo "  help      : Show this help message"

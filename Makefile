GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

BINARY_NAME=bin/ghastly

MAIN_PACKAGE=./cmd

BUILD_FLAGS=-v

# Test flags
TEST_FLAGS=-v


PROTO_DIR=grpc/proto
PROTO_OUT=grpc/gen

PYTHON_OUT=clients/python/gen
NODE_OUT=clients/node/gen
JAVA_OUT=clients/java/src/main/java

SUPPORTED_LANGS=python node java

LANG?=python

.PHONY: all build clean test coverage deps vet fmt lint run help proto validate-lang client-proto

all: deps vet fmt lint coverage build

build:
	GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f ghastly-*

test:
	$(GOTEST) $(TEST_FLAGS) ./tests/...

coverage:
	$(GOTEST) -race -v -covermode=atomic -coverpkg=./...,!/grpc/gen/,!/clients,!/index -coverprofile=coverage.out ./...

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
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o ghastly-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o ghastly-linux-arm64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o ghastly-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o ghastly-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o ghastly-windows-amd64.exe $(MAIN_PACKAGE)


proto:
	mkdir -p $(PROTO_OUT)

	protoc --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		   --go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		   $(PROTO_DIR)/*.proto

validate-lang:
	@if ! echo "$(SUPPORTED_LANGS)" | grep -w "$(LANG)" > /dev/null; then \
		echo "Error: Invalid language '$(LANG)'. Supported languages are: $(SUPPORTED_LANGS)"; \
		exit 1; \
	fi

client-proto: validate-lang
	@echo "Generating $(LANG) client stubs..."
	@mkdir -p $($(shell echo $(LANG) | tr a-z A-Z)_OUT)
	@case "$(LANG)" in \
		"python") \
			python -m grpc_tools.protoc \
				--proto_path=$(PROTO_DIR) \
				--python_out=$(PYTHON_OUT) \
				--grpc_python_out=$(PYTHON_OUT) \
				$(PROTO_DIR)/*.proto ;; \
		"node") \
			grpc_tools_node_protoc \
				--proto_path=$(PROTO_DIR) \
				--js_out=import_style=commonjs:$(NODE_OUT) \
				--grpc_out=grpc_js:$(NODE_OUT) \
				$(PROTO_DIR)/*.proto ;; \
		"java") \
			protoc \
				--proto_path=$(PROTO_DIR) \
				--java_out=$(JAVA_OUT) \
				--grpc-java_out=$(JAVA_OUT) \
				$(PROTO_DIR)/*.proto ;; \
		"go") \
			protoc \
				--proto_path=$(PROTO_DIR) \
				--go_out=$(GO_OUT) --go_opt=paths=source_relative \
				--go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
				$(PROTO_DIR)/*.proto ;; \
	esac
	@echo "Successfully generated $(LANG) client stubs"

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
	@echo "  proto     : Generate protobuf stubs"
	@echo "  run       : Build and run the binary"
	@echo "  build-all : Build binaries for multiple OS and architectures"
	@echo "  help      : Show this help message"
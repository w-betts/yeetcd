# Makefile for yeetcd project
# 
# Targets:
#   build         - Build the yeetcd Go binary
#   test          - Run all tests (Go and Java)
#   test-go       - Run Go tests only
#   test-java     - Run Java tests only
#   clean         - Clean all build artifacts
#   help          - Show this help message

# Default target
.DEFAULT_GOAL := help

# Directories
GOLANG_DIR := golang
JAVA_SDK_DIR := sdks/java
GOPATH := $(shell go env GOPATH)

# Binary output
BINARY_NAME := yeetcd
BINARY_DIR := bin
CLI_DIR := $(BINARY_DIR)/cli

# Platform configurations
PLATFORMS := darwin-amd64 darwin-arm64 linux-amd64 linux-arm64
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Color codes for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Print colored status
define print_status
	@echo "$(GREEN)[✓]$(NC) $(1)"
endef

define print_error
	@echo "$(RED)[✗]$(NC) $(1)"
endef

define print_info
	@echo "$(YELLOW)[→]$(NC) $(1)"
endef

## help - Show this help message
.PHONY: help
help:
	@echo "yeetcd - Build and Test Commands"
	@echo ""
	@echo "Available targets:"
	@grep -E "^[a-zA-Z_-]+:.*##" $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  make %-12s %s\n", $$1, $$2}'

## build - Build the yeetcd Go binary
.PHONY: build
build: $(BINARY_DIR)/$(BINARY_NAME)

$(BINARY_DIR)/$(BINARY_NAME): $(GOLANG_DIR)/cmd/yeetcd
	@mkdir -p $(BINARY_DIR)
	$(call print_info,"Building $(BINARY_NAME)...")
	cd $(GOLANG_DIR) && go build -o ../$(BINARY_DIR)/$(BINARY_NAME) ./cmd/yeetcd
	$(call print_status,"Built $(BINARY_DIR)/$(BINARY_NAME)")

## build-all - Build yeetcd binaries for all platforms (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64)
.PHONY: build-all
build-all: $(foreach platform,$(PLATFORMS),$(CLI_DIR)/$(BINARY_NAME)-$(platform))
	$(call print_status,"Built all platform binaries")

$(CLI_DIR)/$(BINARY_NAME)-%: $(GOLANG_DIR)/cmd/yeetcd
	@mkdir -p $(CLI_DIR)
	$(eval GOOS_ARCH := $(subst -, ,$*))
	$(eval TARGET_GOOS := $(word 1,$(GOOS_ARCH)))
	$(eval TARGET_GOARCH := $(word 2,$(GOOS_ARCH)))
	$(call print_info,"Building $(BINARY_NAME)-$* for $(TARGET_GOOS)/$(TARGET_GOARCH)...")
	cd $(GOLANG_DIR) && GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH) go build -o ../$(CLI_DIR)/$(BINARY_NAME)-$* ./cmd/yeetcd
	$(call print_status,"Built $(CLI_DIR)/$(BINARY_NAME)-$*")

## build-darwin-arm64 - Build yeetcd for darwin arm64
.PHONY: build-darwin-arm64
build-darwin-arm64: $(CLI_DIR)/$(BINARY_NAME)-darwin-arm64

## build-darwin-amd64 - Build yeetcd for darwin amd64
.PHONY: build-darwin-amd64
build-darwin-amd64: $(CLI_DIR)/$(BINARY_NAME)-darwin-amd64

## build-linux-arm64 - Build yeetcd for linux arm64
.PHONY: build-linux-arm64
build-linux-arm64: $(CLI_DIR)/$(BINARY_NAME)-linux-arm64

## build-linux-amd64 - Build yeetcd for linux amd64
.PHONY: build-linux-amd64
build-linux-amd64: $(CLI_DIR)/$(BINARY_NAME)-linux-amd64

## copy-cli-binaries - Copy CLI binaries from bin/cli/ to test resources directory
.PHONY: copy-cli-binaries
copy-cli-binaries:
	$(call print_info,"Copying CLI binaries to test resources...")
	@mkdir -p sdks/java/test/src/main/resources/cli
	@cp -r $(CLI_DIR)/* sdks/java/test/src/main/resources/cli/
	$(call print_status,"CLI binaries copied to sdks/java/test/src/main/resources/cli/")

## test - Run all tests (Go and Java)
.PHONY: test
test: test-go test-java

## test-go - Run Go tests only
.PHONY: test-go
test-go:
	$(call print_info,"Running Go tests...")
	cd $(GOLANG_DIR) && go test -v -race ./...
	$(call print_status,"Go tests passed")

## test-java - Run Java tests only
.PHONY: test-java
test-java:
	$(call print_info,"Running Java tests...")
	cd $(JAVA_SDK_DIR) && ./mvnw clean test
	$(call print_status,"Java tests passed")

## test-go-verbose - Run Go tests with verbose output
.PHONY: test-go-verbose
test-go-verbose:
	cd $(GOLANG_DIR) && go test -v ./...

## install-sdk - Build and install SDK modules to local Maven repository
.PHONY: install-sdk
install-sdk:
	$(call print_info,"Building and installing SDK to local Maven repository...")
	cd $(JAVA_SDK_DIR) && ./mvnw -pl protocol,sdk,test -am clean install
	$(call print_status,"SDK installed to ~/.m2/repository/yeetcd")

## build-sample - Build the sample using locally installed SDK
.PHONY: build-sample
build-sample: install-sdk
	$(call print_info,"Building sample...")
	cd $(JAVA_SDK_DIR)/sample && ../mvnw clean package dependency:copy-dependencies
	$(call print_status,"Sample built")

## test-e2e - Build SDK, run Go e2e tests (requires Docker)
.PHONY: test-e2e
test-e2e: install-sdk
	$(call print_info,"Running E2E tests...")
	cd $(GOLANG_DIR) && go test -v -race -tags=e2e ./e2e/...

## build-java - Build protocol, sdk, and test modules in order
.PHONY: build-java
build-java:
	$(call print_info,"Building Java SDK modules...")
	cd $(JAVA_SDK_DIR) && ./mvnw -pl protocol,sdk,test -am clean compile
	$(call print_status,"Java SDK modules compiled")

## test-sample - Run tests in the sample module
.PHONY: test-sample
test-sample: build-sample
	$(call print_info,"Running sample module tests...")
	cd $(JAVA_SDK_DIR)/sample && ../mvnw test
	$(call print_status,"Sample tests passed")

## test-all - Build everything and run all tests (Go and Java)
.PHONY: test-all
test-all: build-all copy-cli-binaries build-java build-sample test-go test-sample
	$(call print_status,"All tests passed")

## proto - Generate Go code from protobuf definitions
.PHONY: proto
proto:
	$(call print_info,"Generating protobuf Go code...")
	@mkdir -p $(GOLANG_DIR)/internal/core/proto/pipeline
	@mkdir -p $(GOLANG_DIR)/internal/core/proto/mock
	PATH="$(GOPATH)/bin:$(PATH)" protoc --go_out=$(GOLANG_DIR)/internal/core/proto/pipeline --go_opt=paths=source_relative \
		--go-grpc_out=$(GOLANG_DIR)/internal/core/proto/pipeline --go-grpc_opt=paths=source_relative \
		--proto_path=$(GOLANG_DIR)/protocol/src/main/proto \
		$(GOLANG_DIR)/protocol/src/main/proto/yeetcd/protocol/pipeline/pipeline.proto
	PATH="$(GOPATH)/bin:$(PATH)" protoc --go_out=$(GOLANG_DIR)/internal/core/proto/mock --go_opt=paths=source_relative \
		--go-grpc_out=$(GOLANG_DIR)/internal/core/proto/mock --go-grpc_opt=paths=source_relative \
		--proto_path=$(GOLANG_DIR)/protocol/src/main/proto \
		$(GOLANG_DIR)/protocol/src/main/proto/yeetcd/protocol/mock/mock.proto
	$(call print_status,"Protobuf code generated")
	$(call print_status,"E2E tests passed")

## clean - Clean all build artifacts
.PHONY: clean
clean:
	$(call print_info,"Cleaning build artifacts...")
	rm -rf $(BINARY_DIR)
	cd $(GOLANG_DIR) && go clean ./...
	cd $(JAVA_SDK_DIR) && ./mvnw clean
	$(call print_status,"Cleaned all artifacts")

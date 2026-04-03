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

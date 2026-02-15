# Makefile for chkiso
# Builds binaries for multiple platforms

VERSION := 2.0.0
BINARY_NAME := chkiso
BUILD_DIR := build

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -s -w
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

.PHONY: all build clean test windows linux macos darwin build-all

# Default target
all: build

# Build for current platform
build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for all platforms
build-all: windows linux macos

# Windows builds
windows: windows-amd64 windows-arm64

windows-amd64:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS) -H windowsgui" -trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe

windows-arm64:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe

# Linux builds
linux: linux-amd64 linux-386 linux-arm64 linux-arm

linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64

linux-386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386

linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64

linux-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm

# macOS builds
macos: darwin

darwin: darwin-amd64 darwin-arm64

darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64

darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64

# FreeBSD builds (bonus)
freebsd: freebsd-amd64 freebsd-arm64

freebsd-amd64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-freebsd-amd64

freebsd-arm64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-freebsd-arm64

# Install locally
install: build
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Uninstall
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

# Show help
help:
	@echo "Available targets:"
	@echo "  make build       - Build for current platform"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make windows     - Build for Windows (all architectures)"
	@echo "  make linux       - Build for Linux (all architectures)"
	@echo "  make macos       - Build for macOS (all architectures)"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make install     - Install binary locally"
	@echo "  make uninstall   - Uninstall binary"

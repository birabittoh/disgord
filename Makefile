# Variables
APP_NAME=disgord
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_DIR=dist
SRC_DIR=src

# Build flags for versioning
LDFLAGS=-ldflags "-X github.com/birabittoh/disgord/src/globals.CommitID=$(COMMIT_HASH)"

.PHONY: all build test run clean

# Default command: build the application
all: build

# Build the Go application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(APP_NAME)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run .

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)

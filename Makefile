# Makefile for Ttracker

# Variables
BINARY_NAME=tt
GO=go

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	$(GO) build -o $(BINARY_NAME) main.go

# Clean up binary
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	@echo "Cleaned up $(BINARY_NAME)"

# Build and run the daemon
.PHONY: daemon
daemon: build
	./$(BINARY_NAME) daemon

# Build and start the daemon (alternative name)
.PHONY: start
start: build
	@echo "Starting Ttracker daemon..."
	./$(BINARY_NAME) daemon

# Help
.PHONY: help
help:
	@echo "Ttracker Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build    - Build the application"
	@echo "  make clean    - Remove the binary"
	@echo "  make daemon   - Build and run the daemon"
	@echo "  make start    - Build and start the daemon (same as daemon)"
	@echo "  make help     - Show this help message" 
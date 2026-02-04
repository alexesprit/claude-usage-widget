.PHONY: build clean test release

# Default build (optimized)
build:
	go build -ldflags="-s -w" -o usage .

# Development build (with debug info)
dev:
	go build -o usage .

# Release build (optimized)
release: build
	@echo "Release build complete!"

# Clean build artifacts
clean:
	rm -f usage usage_compressed

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build optimized binary (default)"
	@echo "  dev      - Build with debug info"
	@echo "  release  - Build optimized release binary"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  help     - Show this help"
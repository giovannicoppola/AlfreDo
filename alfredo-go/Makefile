.PHONY: build clean test run install deps

# Build the application
build:
	go build -o alfredo-go main.go

# Build for different platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o alfredo-go-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o alfredo-go-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -o alfredo-go-linux-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o alfredo-go-windows-amd64.exe main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test ./...

# Run the application (requires TOKEN env var)
run:
	go run main.go

# Clean build artifacts
clean:
	rm -f alfredo-go*

# Install to local bin (optional)
install: build
	cp alfredo-go /usr/local/bin/

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  build-all - Build for multiple platforms"
	@echo "  deps      - Install dependencies"
	@echo "  test      - Run tests"
	@echo "  run       - Run the application"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install to /usr/local/bin"
	@echo "  fmt       - Format code"
	@echo "  lint      - Lint code"
	@echo "  help      - Show this help"

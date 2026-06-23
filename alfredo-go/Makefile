.PHONY: build clean test run install deps build-universal build-workflow sign release

# Code signing / notarization config (override on the command line if needed)
SIGN_IDENTITY  ?= C85516732AD930BB84E217AFF2C375E64268B111
NOTARY_PROFILE ?= gio-notarytool
WORKFLOW_DIR   ?= ../source
BINARY         ?= alfredo-go

# Build the application
build:
	go build -o alfredo-go main.go

# Build universal macOS binary (amd64 + arm64)
build-universal:
	GOOS=darwin GOARCH=amd64 go build -o alfredo-go-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o alfredo-go-arm64 main.go
	lipo -create -output alfredo-go alfredo-go-amd64 alfredo-go-arm64
	rm -f alfredo-go-amd64 alfredo-go-arm64

# Build and copy binary into the Alfred workflow directory
build-workflow: build-universal
	cp alfredo-go ../source/alfredo-go

# Sign and notarize the binary in the Alfred workflow directory
sign:
	cd $(WORKFLOW_DIR) && \
		codesign -s $(SIGN_IDENTITY) -o runtime -v $(BINARY) && \
		zip $(BINARY).zip $(BINARY) && \
		xcrun notarytool submit $(BINARY).zip --keychain-profile "$(NOTARY_PROFILE)" --wait --verbose && \
		rm -f $(BINARY).zip && \
		codesign --display --verbose=4 $(BINARY)

# Full release: build universal binary, copy into workflow, then sign + notarize
release: build-workflow sign

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
	@echo "  build           - Build the application"
	@echo "  build-universal - Build universal macOS binary (amd64+arm64)"
	@echo "  build-workflow  - Build and copy binary into Alfred workflow"
	@echo "  sign            - Sign + notarize the binary in the workflow dir"
	@echo "  release         - build-workflow then sign + notarize"
	@echo "  build-all       - Build for multiple platforms"
	@echo "  deps            - Install dependencies"
	@echo "  test            - Run tests"
	@echo "  run             - Run the application"
	@echo "  clean           - Clean build artifacts"
	@echo "  install         - Install to /usr/local/bin"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  help            - Show this help"

.PHONY: build run test clean lint

# Build the application binary
build:
	go build -o bin/xstream-tui ./cmd/xstream-tui

# Run the application directly
run:
	go run ./cmd/xstream-tui

# Run all tests with verbose output
test:
	go test -v ./...

# Run tests with coverage report
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run linter
lint:
	go vet ./...

# Tidy dependencies
tidy:
	go mod tidy

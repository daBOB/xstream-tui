.PHONY: build run test clean lint docker-image docker-run docker-down docker-logs vpn-check

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

# --- Docker + Mullvad kill switch ---

# Build the container image
docker-image:
	docker compose build

# Start the VPN and attach the TUI to it (interactive)
docker-run:
	docker compose run --rm xstream-tui

# Stop the VPN container
docker-down:
	docker compose down

# Follow gluetun tunnel logs
docker-logs:
	docker compose logs -f gluetun

# Verify the app container really exits through Mullvad
vpn-check:
	docker compose run --rm --entrypoint sh xstream-tui -c \
		'wget -qO- https://am.i.mullvad.net/json'

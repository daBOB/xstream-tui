.PHONY: build run test clean lint docker-image docker-run docker-down docker-logs vpn-check vpn-up vpn-down tui

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

# --- Plain docker run (same stack without compose) ---

DOWNLOAD_DIR ?= $(HOME)/TK08/xstream
CONFIG_DIR   ?= $(CURDIR)/data/config

# Start the Mullvad tunnel container
vpn-up:
	docker volume create gluetun-data
	docker run -d --name xstream-gluetun \
		--cap-add NET_ADMIN \
		--device /dev/net/tun:/dev/net/tun \
		--env-file docker/mullvad.env \
		-e VPN_SERVICE_PROVIDER=mullvad \
		-e VPN_TYPE=wireguard \
		-e FIREWALL=on \
		-e DOT=on \
		-e HEALTH_VPN_DURATION_INITIAL=20s \
		-e TZ=Europe/Berlin \
		-v gluetun-data:/gluetun \
		--restart unless-stopped \
		qmcgaw/gluetun:v3
	@echo "waiting for the tunnel..."
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' xstream-gluetun)" = healthy ]; do sleep 3; done
	@echo "tunnel healthy"

# Stop and remove the tunnel container
vpn-down:
	docker rm -f xstream-gluetun

# Run the TUI in the tunnel's network namespace (needs a real terminal)
tui:
	docker run --rm -it --name xstream-tui \
		--network container:xstream-gluetun \
		--env-file .env \
		-e TERM=$(TERM) \
		-v "$(CONFIG_DIR):/config" \
		-v "$(DOWNLOAD_DIR):/downloads" \
		xstream-tui:local

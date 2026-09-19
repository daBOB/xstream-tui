# syntax=docker/dockerfile:1

# ---------- build stage ----------
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Download modules first so source edits do not invalidate the dependency layer
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# CGO_ENABLED=0 produces a fully static binary, so the runtime stage needs no libc
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
      -o /out/xstream-tui ./cmd/xstream-tui

# ---------- runtime stage ----------
FROM alpine:3.22

# ca-certificates: TLS to the Xtream portal
# ncurses-terminfo-base: key and colour handling for the Bubble Tea TUI
# tzdata: correct local timestamps for account expiry rendering
RUN apk add --no-cache ca-certificates ncurses-terminfo-base tzdata \
    && adduser -D -u 1000 -h /home/xstream xstream \
    && mkdir -p /config /downloads \
    && chown xstream:xstream /config /downloads

COPY --from=builder /out/xstream-tui /usr/local/bin/xstream-tui

# No mpv/vlc in this image by design: playback stays on the host, only the
# portal API and downloads run inside the VPN network namespace.
ENV XDG_CONFIG_HOME=/config \
    XSTREAM_DOWNLOAD_DIR=/downloads \
    TERM=xterm-256color

USER xstream
WORKDIR /home/xstream
VOLUME ["/config", "/downloads"]

ENTRYPOINT ["xstream-tui"]

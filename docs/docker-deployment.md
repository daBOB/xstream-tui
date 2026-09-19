# Docker Deployment with Mullvad Kill Switch

Runs xstream-tui inside a container whose only route to the internet is a
Mullvad WireGuard tunnel. If the tunnel drops, the container loses network
access instead of falling back to your normal connection.

## What runs where

| Concern | Location | Through VPN |
|---|---|---|
| Xtream portal API (login, categories, streams) | container | yes |
| Downloads | container → `~/TK08/xstream` | yes |
| Playback (mpv / VLC) | host, on the downloaded files | n/a |

The image deliberately contains **no media player**. Playback needs a display
server and GPU access, which would mean punching X11/Wayland sockets and
`/dev/dri` into the container. Starting playback from inside the container
shows the app's normal "no player available" error; play the finished files
with your host mpv/VLC from `~/TK08/xstream`.

## How the kill switch works

Two containers, one network namespace:

```
┌─ gluetun ──────────────────────────────────┐
│  WireGuard → Mullvad                       │
│  nftables: default DROP, allow only tun0   │
│  DNS-over-TLS (no LAN resolver leak)       │
│                    ▲                       │
│   xstream-tui ─────┘  network_mode:        │
│   (no netstack of its own) service:gluetun │
└────────────────────────────────────────────┘
```

Three independent failure modes all fail closed:

1. **Tunnel drops** → gluetun's firewall (`FIREWALL=on`) drops every packet
   that is not the tunnel or the WireGuard endpoint.
2. **gluetun container stops** → its network namespace disappears, and the app
   container has no interface at all.
3. **gluetun never comes up** → `depends_on: condition: service_healthy` stops
   the app container from starting in the first place.

## Setup

1. Get a WireGuard config from
   <https://mullvad.net/en/account/wireguard-config>: log in with your account
   number → **Generate key** (one key per device; reusing a key across devices
   causes connectivity problems) → pick country/city/server → optionally enable
   DNS content blockers → **Download file**. You get `mullvad-<server>.conf`.

   Mullvad is WireGuard-only since [OpenVPN was removed on 15 January
   2026](https://mullvad.net/en/blog/removing-openvpn-15th-january-2026),
   client *and* server side — hence `VPN_TYPE: wireguard` in the compose file.

   The private key and the tunnel address are the same for every Mullvad
   server, so switching exit servers later means changing only the server
   selection. Regenerating the key changes the tunnel address with it.

2. Convert it to the env file gluetun reads:

   ```bash
   scripts/mullvad-wg-to-env.sh ~/Downloads/mullvad-de-ber-wg-001.conf
   ```

   This writes `docker/mullvad.env` (mode 600, gitignored) with
   `WIREGUARD_PRIVATE_KEY`, `WIREGUARD_ADDRESSES` and `SERVER_HOSTNAMES`.
   To change servers later, edit `SERVER_HOSTNAMES` (or swap it for
   `SERVER_CITIES=Berlin` / `SERVER_COUNTRIES=Germany`) — the key stays valid.

3. Optional: put portal credentials in `.env` at the repo root
   (`XSTREAM_HOST`, `XSTREAM_PORT`, `XSTREAM_USERNAME`, `XSTREAM_PASSWORD`) to
   prefill the login screen. It is mounted as an env file, not baked into the
   image.

4. Build and run:

   ```bash
   make docker-image     # docker compose build
   make docker-run       # docker compose run --rm xstream-tui
   ```

   `docker compose run` is required rather than `up`, because a TUI needs an
   attached TTY.

## Verifying the tunnel

```bash
make vpn-check
```

Runs `wget -qO- https://am.i.mullvad.net/json` *from inside the app container*.
`"mullvad_exit_ip": true` plus the expected exit country means traffic really
leaves through Mullvad.

To prove the kill switch, break the tunnel and re-check. Both failure modes were
verified on this setup:

```bash
# 1. gluetun stopped -> the app container cannot even start
docker compose stop gluetun
make vpn-check
# -> "cannot join network namespace of a non running container"

# 2. gluetun up but tunnel down -> firewall drops everything
docker compose start gluetun
docker exec xstream-gluetun ip link set tun0 down
make vpn-check
# -> "wget: bad address 'am.i.mullvad.net'" (DNS dies with the tunnel; no leak)

docker compose up -d --force-recreate gluetun   # restore
```

## Data and permissions

| Path in container | Host path | Contents |
|---|---|---|
| `/config` | `./data/config` | `config.toml`, `credentials.json` |
| `/downloads` | `~/TK08/xstream` (SSHFS share) | downloaded media |

The container runs as uid 1000. If your user has a different uid, either
`chown -R 1000:1000 data/` or change the `user:` line in `docker-compose.yml`.

## LAN access

The kill switch blocks the LAN too. To let the container write to a NAS or
reach a local server, list the subnet in `docker/mullvad.env`:

```
FIREWALL_OUTBOUND_SUBNETS=192.168.1.0/24
```

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `failed to add the host (vethXXXX) <=> sandbox pair interfaces: operation not supported` | Kernel was upgraded without rebooting, so the `veth` module for the running kernel is gone. Reboot. |
| gluetun logs `wireguard private key is not set` | `docker/mullvad.env` still has the example placeholders — run `scripts/mullvad-wg-to-env.sh`. |
| gluetun stuck `unhealthy`, logs repeat `i/o timeout` on DNS, and `tun0` shows TX > 0 but **RX 0** | Handshake never answered. On this network UDP 51820 is blocked upstream — Mullvad also listens on UDP 53, so set `WIREGUARD_ENDPOINT_PORT=53`. Rule out a dead server first by switching `SERVER_HOSTNAMES` to `SERVER_COUNTRIES`; if several servers are all silent, it is the port. |
| gluetun exits with `interface address is IPv6 but IPv6 is not supported` | Docker has no IPv6 by default. `WIREGUARD_ADDRESSES` must carry only the IPv4 address; the conversion script strips the `fc00:bbbb:…` one automatically. |
| App container never starts | gluetun is unhealthy; `make docker-logs` shows the handshake failure. A rotated/revoked Mullvad key is the usual cause. |
| Garbled TUI output | Pass your terminal through: `TERM=$TERM docker compose run --rm xstream-tui`. |
| Permission denied writing downloads | The download target must be writable by uid 1000. For the SSHFS share, check the mount is up (`df -h ~/TK08`) and exposes itself to root — a FUSE mount without `allow_other` is invisible to the Docker daemon and the bind mount comes up empty. |

## Files

- `Dockerfile` — two-stage static Go build → Alpine runtime, non-root
- `docker-compose.yml` — gluetun + app, shared netns
- `docker/mullvad.env.example` — template for the VPN secrets
- `scripts/mullvad-wg-to-env.sh` — Mullvad `.conf` → env file
- `.dockerignore` — keeps `.env`, `docker/mullvad.env` and keys out of the build context

# wa-gateway-go

[![CI](https://github.com/AsyrafHussin/wa-gateway-go/actions/workflows/ci.yml/badge.svg)](https://github.com/AsyrafHussin/wa-gateway-go/actions/workflows/ci.yml)
[![Release](https://github.com/AsyrafHussin/wa-gateway-go/actions/workflows/release.yml/badge.svg)](https://github.com/AsyrafHussin/wa-gateway-go/actions/workflows/release.yml)

A lightweight, self-hosted WhatsApp gateway built with Go. Single binary, no dependencies, low memory footprint.

Connect WhatsApp devices via QR code or pairing code, send messages, validate phone numbers, and capture contacts — all through a simple REST API with real-time WebSocket events.

## Features

- **Single binary** — no runtime dependencies, just copy and run
- **Multi-device** — manage multiple WhatsApp accounts simultaneously
- **QR code & pairing code** — two ways to link devices
- **Contact capture** — automatically collect contacts from incoming messages and history sync
- **Real-time events** — native WebSocket for QR codes, connection status, and more
- **Webhooks** — HTTP callbacks with HMAC-SHA256 signing for message receipts and device events
- **Phone validation** — check if numbers are registered on WhatsApp with built-in caching
- **Low memory** — ~30-80 MB per connected device
- **Cross-platform** — builds for Linux, macOS, and Windows

## Quick Start

### Download

Grab a pre-built binary from [Releases](https://github.com/AsyrafHussin/wa-gateway-go/releases), or build from source:

```bash
# Clone and build
git clone https://github.com/AsyrafHussin/wa-gateway-go.git
cd wa-gateway-go
go build -o wa-gateway-go .
```

### Configure

```bash
cp .env.example .env
# Edit .env and set your API_KEY
```

### Run

```bash
./wa-gateway-go
```

The server starts on `http://0.0.0.0:4010` by default.

## Build

```bash
# Local build
make build

# Linux (for VPS deployment)
make build-linux

# Development (auto-run)
make dev
```

## Code Quality

```bash
# Format all files (goimports — like Prettier for Go)
make fmt

# Lint all files (golangci-lint — like ESLint for Go)
make lint
```

The project uses [golangci-lint](https://golangci-lint.run/) with errcheck, govet, staticcheck, unused, and ineffassign linters. See `.golangci.yml` for configuration.

## Configuration

All configuration is via environment variables or a `.env` file.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `4010` | Server port |
| `HOST` | `0.0.0.0` | Server bind address |
| `API_KEY` | — | **Required.** API key for authentication |
| `LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `CORS_ORIGINS` | `*` | Allowed CORS origins |
| `PHONE_COUNTRY_CODE` | `60` | Default country code for phone validation |
| `PHONE_MIN_LENGTH` | `11` | Minimum phone number length (with country code) |
| `PHONE_MAX_LENGTH` | `12` | Maximum phone number length (with country code) |
| `DATA_DIR` | `./data` | Directory for session and contact databases |
| `TYPING_DELAY_MS` | `1000` | Simulated typing delay before sending messages |
| `AUTO_READ_RECEIPT` | `false` | Automatically mark incoming messages as read |
| `WEBHOOK_URL` | — | URL to receive webhook events |
| `WEBHOOK_SECRET` | — | Secret for HMAC-SHA256 webhook signatures |
| `WEBHOOK_TIMEOUT_MS` | `5000` | Webhook request timeout |
| `RATE_LIMIT_DEVICES` | `10` | Device endpoints: requests per minute |
| `RATE_LIMIT_MESSAGES` | `30` | Message endpoints: requests per minute |
| `RATE_LIMIT_VALIDATE` | `60` | Validation endpoints: requests per minute |
| `CACHE_TTL_SECONDS` | `3600` | Phone validation cache TTL |
| `WS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed WebSocket origins |
| `WS_AUTH_TIMEOUT` | `5` | Seconds to wait for WebSocket auth (1-60) |

## Authentication

All API endpoints (except `/health`) require authentication. REST endpoints use header-based auth, WebSocket uses first-message auth.

**REST endpoints** — authenticate via one of:

```
Authorization: Bearer YOUR_API_KEY
```

```
X-API-Key: YOUR_API_KEY
```

## API Reference

See [API.md](API.md) for the full API documentation with request/response examples.

### Endpoints Overview

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | No | Basic health check |
| `GET` | `/health/detailed` | Yes | Detailed health with memory and device stats |
| `POST` | `/devices` | Yes | Connect a WhatsApp device |
| `DELETE` | `/devices/:token` | Yes | Disconnect and logout a device |
| `POST` | `/messages` | Yes | Send a text message |
| `POST` | `/validate/phone` | Yes | Check if a phone number is on WhatsApp |
| `GET` | `/contacts/:token` | Yes | List captured contacts for a device |
| `DELETE` | `/cache` | Yes | Clear phone validation cache |
| `GET` | `/ws` | WS Auth | WebSocket for real-time events |

## WebSocket Events

Connect to `ws://localhost:4010/ws` and authenticate with a first-message API key exchange:

```js
const ws = new WebSocket('ws://localhost:4010/ws');
ws.onopen = () => ws.send(JSON.stringify({ type: 'auth', apiKey: 'YOUR_API_KEY' }));
```

**Security:** Origin whitelist (`WS_ALLOWED_ORIGINS`) blocks cross-site hijacking at upgrade time. First-message auth with constant-time key comparison blocks unauthorized access. Unauthenticated clients receive no broadcasts.

| Event | Description |
|---|---|
| `qrcode` | QR code string for scanning |
| `pairing-code` | Pairing code for phone-based linking |
| `connection-success` | Device connected successfully |
| `connection-error` | Connection failed or device disconnected |

### Message Format

```json
{
  "event": "qrcode",
  "token": "60123456789",
  "data": "2@ABC123..."
}
```

## Webhooks

When `WEBHOOK_URL` is configured, the gateway sends HTTP POST requests for device and message events.

| Event | Description |
|---|---|
| `device.connected` | Device connected to WhatsApp |
| `device.disconnected` | Device disconnected |
| `device.logged_out` | Device was logged out (needs re-pairing) |
| `message.receipt` | Message delivery/read receipt |
| `contacts.new` | New contact captured from incoming message |
| `contacts.sync` | Batch contacts from history sync |

### Signature Verification

If `WEBHOOK_SECRET` is set, each request includes an `X-Webhook-Signature` header containing an HMAC-SHA256 hex digest of the request body:

```
X-Webhook-Signature: sha256=<hex_digest>
```

## Project Structure

```
wa-gateway-go/
├── main.go                     # Entry point, dependency wiring, graceful shutdown
├── config/
│   └── config.go               # Environment variable parsing and validation
├── internal/
│   ├── server/server.go        # Fiber app setup, CORS, route registration
│   ├── middleware/
│   │   ├── auth.go             # Bearer token / X-API-Key authentication
│   │   ├── ratelimit.go        # Fixed-window per-IP rate limiter
│   │   └── logger.go           # Structured request logging
│   ├── handler/                # HTTP request handlers
│   ├── whatsapp/
│   │   ├── manager.go          # Multi-device session manager
│   │   ├── session.go          # WhatsApp client lifecycle
│   │   ├── events.go           # Event dispatcher
│   │   ├── sender.go           # Message sending with typing simulation
│   │   └── contacts.go         # Contact extraction from messages/history
│   ├── ws/                     # WebSocket hub and client management
│   ├── webhook/webhook.go      # Webhook dispatcher with HMAC signing
│   ├── contacts/store.go       # SQLite-backed contact storage
│   └── cache/cache.go          # In-memory phone validation cache
├── pkg/
│   ├── response/response.go    # JSON response envelope helpers
│   └── validator/validator.go  # Phone number and message validation
└── data/                       # Runtime data (auto-created)
    ├── sessions/               # WhatsApp session databases (per device)
    └── contacts/               # Contact databases (per device)
```

## Tech Stack

| Component | Package |
|---|---|
| HTTP framework | [Fiber v2](https://github.com/gofiber/fiber) |
| WhatsApp client | [whatsmeow](https://github.com/tulir/whatsmeow) |
| WebSocket | [gofiber/contrib/websocket](https://github.com/gofiber/contrib/tree/main/websocket) |
| SQLite | [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) (pure Go, no CGo) |
| Logging | [zerolog](https://github.com/rs/zerolog) |
| Cache | [go-cache](https://github.com/patrickmn/go-cache) |
| Config | [godotenv](https://github.com/joho/godotenv) |

## Deployment

### Systemd Service

```ini
[Unit]
Description=WhatsApp Gateway
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/wa-gateway-go
ExecStart=/opt/wa-gateway-go/wa-gateway-go
Restart=always
RestartSec=5
EnvironmentFile=/opt/wa-gateway-go/.env

[Install]
WantedBy=multi-user.target
```

### Deploy Steps

```bash
# Build for Linux
make build-linux

# Copy to server
scp wa-gateway-go .env user@server:/opt/wa-gateway-go/

# On server
sudo systemctl enable wa-gateway-go
sudo systemctl start wa-gateway-go
```

## Roadmap

- [ ] Media messages (image, video, audio, document)
- [ ] Group messaging
- [ ] Reactions, replies, and read receipts
- [ ] Message edit and delete
- [ ] Polls
- [ ] Newsletter/Channel support
- [ ] Presence tracking (online/offline)
- [ ] Auto-reject calls
- [ ] Docker image

## License

[MIT](LICENSE)

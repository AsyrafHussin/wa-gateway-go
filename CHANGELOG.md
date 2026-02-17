# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-17

### Added

- **HTTP API** — Fiber v2 server with structured JSON responses, CORS, and panic recovery
- **Authentication** — Bearer token and X-API-Key header support with constant-time comparison
- **Rate limiting** — Fixed-window per-IP rate limiter, configurable per endpoint group
- **Request logging** — Structured JSON logs via zerolog (method, path, status, latency, IP)
- **Device management** — Connect via QR code or pairing code, disconnect with full logout
- **Text messaging** — Send messages with configurable typing delay simulation
- **Phone validation** — Check if numbers are on WhatsApp, with in-memory TTL cache
- **Contact capture** — Automatically extract contacts from incoming messages
- **History sync** — Capture contacts from WhatsApp history on first connect
- **Contact export** — List contacts as JSON (paginated) or CSV download
- **WebSocket hub** — Real-time events for QR codes, pairing codes, and connection status
- **Webhooks** — HTTP POST callbacks with HMAC-SHA256 signing for device and message events
- **Auto-reconnect** — Existing sessions reconnect automatically on server restart
- **Graceful shutdown** — Clean disconnect of all devices on SIGINT/SIGTERM
- **Pure Go SQLite** — No CGo dependency, cross-compiles to static Linux binary
- **Configurable** — All settings via environment variables or `.env` file

### Webhook Events

- `device.connected` / `device.disconnected` / `device.logged_out`
- `message.receipt` (delivered, read, played)
- `contacts.new` (from incoming messages)
- `contacts.sync` (from history sync)

### WebSocket Events

- `qrcode` — QR code string for rendering
- `pairing-code` — Phone-based pairing code
- `connection-success` — Device paired and connected
- `connection-error` — Connection failed or lost

[0.1.0]: https://github.com/AsyrafHussin/wa-gateway-go/releases/tag/v0.1.0

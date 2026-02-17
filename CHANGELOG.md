# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.5] - 2026-02-17

### Added

- **Docker support** — multi-stage Dockerfile (golang:1.25-alpine → alpine:3.21) with non-root user, ~15MB image
- **docker-compose.yml** — single service with named volume for persistent WhatsApp session data
- **setup.sh** — interactive `.env` setup: auto-generate (random API_KEY + defaults) or manual entry
- **Makefile targets** — `make setup`, `make docker`, `make docker-down`, `make docker-logs`
- **.dockerignore** — excludes .git, data/, binaries, .env, docs, IDE files

### Changed

- README Deployment section expanded with 4 copy-paste-ready options: Docker Compose, Docker, pre-built binary, build from source

## [0.1.4] - 2026-02-17

### Added

- **WebSocket origin whitelist** — `WS_ALLOWED_ORIGINS` env var blocks cross-site WebSocket hijacking (CSWSH) at HTTP upgrade time
- **WebSocket first-message auth** — clients must send `{"type":"auth","apiKey":"..."}` within configurable timeout (`WS_AUTH_TIMEOUT`, default 5s)
- **Broadcast filtering** — only authenticated WebSocket clients receive broadcast messages
- **Security event logging** — auth failures, timeouts, and origin rejections logged with remote IP at Warn level
- **Read size limit** — `conn.SetReadLimit(4096)` prevents OOM from oversized WebSocket messages
- **Config validation** — `WS_AUTH_TIMEOUT` validated to 1-60 range
- **56 new tests across 6 test files:**
  - `config/config_test.go` (13 tests) — defaults, custom values, port/timeout validation, env helpers
  - `internal/ws/hub_test.go` (15 tests) — register/unregister, broadcast to authenticated clients, data types, shutdown, skip unauthenticated, none authenticated
  - `internal/ws/auth_test.go` (5 tests) — default auth state, auth flag, auth message/response structs, timeout cancellation
  - `internal/ws/security_test.go` (4 tests) — constant-time comparison, null bytes, replay auth (sync.Once), oversized payload limit
  - `internal/handler/ws_test.go` (10 tests) — origin allowed/rejected/wildcard/empty/multiple/case-insensitive, empty config defaults to wildcard, whitespace config defaults to wildcard, config parsing
  - `internal/handler/ws_integration_test.go` (9 tests) — full WebSocket auth flow: valid key, invalid key, malformed JSON, wrong type, empty key, timeout, broadcast after auth, origin rejected, origin allowed

### Fixed

- **Auth timeout goroutine leak** — `authDone` channel now closed on all auth exit paths, `AuthTimeoutPump` exits immediately
- **Auth timeout race** — `authDone` closed before setting `authenticated` flag, preventing post-auth disconnection

### Security

- Constant-time API key comparison (`crypto/subtle.ConstantTimeCompare`)
- Null byte and oversized payload rejection
- Empty/whitespace origin config safely defaults to wildcard

## [0.1.3] - 2026-02-17

### Added

- **CI workflow** — lint, build, and test on every push/PR to main
- **Release workflow** — auto-build binaries (Linux amd64/arm64, macOS amd64/arm64) and attach to GitHub releases on tag push
- CI/CD status badges in README
- Pre-built binary download link in Quick Start

## [0.1.2] - 2026-02-17

### Added

- **golangci-lint config** — `.golangci.yml` with errcheck, govet, staticcheck, unused, ineffassign linters
- **goimports formatter** — auto-groups imports into stdlib, third-party, and local sections
- **`make fmt`** — format all Go files with goimports
- **`make lint`** — run golangci-lint across the entire codebase

### Fixed

- 20 unchecked error returns (`errcheck`) across 11 files
- Incorrect `int`-to-`string` conversion in TemporaryBan event logging (`govet`)
- Import ordering to separate third-party from local imports

## [0.1.1] - 2026-02-17

### Fixed

- **Race condition in WebSocket hub** — broadcast now uses write lock when removing slow clients
- **CSV injection** — contact CSV export uses `encoding/csv` writer instead of string concatenation
- **Path traversal** — token parameters validated as digits-only (7-15 chars) across all endpoints
- **Internal error exposure** — error responses no longer leak internal error messages to clients
- **Context cancellation** — typing delay in message sender now respects context cancellation
- **Resource leak on connect failure** — failed device connections are properly cleaned up from session map
- **WebSocket handler allocation** — handler created once at init instead of per-request
- **History sync webhook payload** — sends only newly discovered contacts instead of entire contact store
- **Missing rows.Err() check** — contact store query now checks for iteration errors
- **Directory creation errors** — session/contact directory creation failures are now surfaced
- **Logout error handling** — WhatsApp logout errors are now logged instead of silently discarded
- **Negative pagination** — contact list endpoint rejects negative limit/offset values
- **Hub shutdown** — WebSocket hub now has proper shutdown mechanism, wired into graceful shutdown

### Changed

- Deduplicated WebSocket broadcast methods into shared internal helper
- Added `INVALID_TOKEN` error code to API documentation

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

[0.1.5]: https://github.com/AsyrafHussin/wa-gateway-go/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/AsyrafHussin/wa-gateway-go/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/AsyrafHussin/wa-gateway-go/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/AsyrafHussin/wa-gateway-go/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/AsyrafHussin/wa-gateway-go/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/AsyrafHussin/wa-gateway-go/releases/tag/v0.1.0

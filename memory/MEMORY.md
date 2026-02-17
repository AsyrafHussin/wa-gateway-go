# wa-gateway-go Memory

## Architecture
- Go binary with Fiber v2 HTTP, native WebSocket, whatsmeow WhatsApp client
- Pure Go SQLite (modernc.org/sqlite) — no CGo, cross-compiles cleanly
- Multi-device: DeviceManager holds map[token]*DeviceSession

## Key Files
- `main.go` — entry point, wiring, graceful shutdown
- `config/config.go` — env parsing with defaults
- `internal/whatsapp/manager.go` — DeviceManager (thread-safe session map)
- `internal/whatsapp/session.go` — DeviceSession (whatsmeow client lifecycle)
- `internal/whatsapp/events.go` — event handler (type-switch)
- `internal/whatsapp/sender.go` — SendText with typing simulation
- `internal/ws/hub.go` — WebSocket broadcast hub
- `internal/webhook/webhook.go` — HTTP POST dispatcher with HMAC-SHA256

## NOW
- Initial implementation complete, needs real device testing
- Binary builds: `go build` (macOS), `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build` (Linux)

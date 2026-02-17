# Session: 2026-02-17 (WS Security)

## Completed Tasks
- WebSocket security: origin whitelist + first-message auth — 06d067e
  - Layer 1: WS_ALLOWED_ORIGINS origin check at HTTP upgrade time
  - Layer 2: First-message auth with constant-time API key comparison
  - Broadcast filtering: only authenticated clients receive messages
  - Security event logging with remote IP (zerolog Warn level)
  - conn.SetReadLimit(4096) prevents OOM
  - WSAuthTimeout config validation (1-60s)
  - authDone goroutine leak fix + auth timeout race fix
  - 56 new tests across 6 test files
  - Updated README, API.md, CHANGELOG, .env.example
- Lint fixes for errcheck/goimports — pending commit

## Decisions
- Used `t.Setenv()` instead of `os.Setenv` in tests (auto-cleanup, no errcheck issues)
- Default WS_ALLOWED_ORIGINS=* (matches CORS_ORIGINS default, dev-friendly)
- Empty/whitespace origins config defaults to wildcard (safety net)
- Single API key shared for REST + WS auth (same key, different transport)

## Gotchas
- `errcheck` linter requires `_ =` on all `os.Unsetenv`, `conn.Close()`, `conn.SetReadDeadline`
- `goimports` enforces comment alignment in struct literals
- CI lint runs golangci-lint v2 — must pass locally before push
- `authDone` must be closed on ALL exit paths (success AND failure) to prevent goroutine leak
- Close `authDone` BEFORE `authenticated.Store(true)` to prevent timeout race

## Next Steps
- v0.1.4 tag pushed, CI/CD release building binaries
- Test with real WhatsApp device
- Consider WS rate limiting / max connections for production hardening

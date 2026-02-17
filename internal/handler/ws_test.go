package handler

import (
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
)

func newTestWSHandler(allowedOrigins string) *WS {
	logger := zerolog.New(io.Discard)
	hub := ws.NewHub(logger)
	return NewWS(hub, "test-api-key", 5*time.Second, allowedOrigins, logger)
}

func TestWS_OriginAllowed(t *testing.T) {
	h := newTestWSHandler("http://localhost:3000,https://app.example.com")

	if !h.isOriginAllowed("http://localhost:3000") {
		t.Error("expected http://localhost:3000 to be allowed")
	}
	if !h.isOriginAllowed("https://app.example.com") {
		t.Error("expected https://app.example.com to be allowed")
	}
}

func TestWS_OriginRejected(t *testing.T) {
	h := newTestWSHandler("http://localhost:3000")

	if h.isOriginAllowed("http://evil.com") {
		t.Error("expected http://evil.com to be rejected")
	}
	if h.isOriginAllowed("https://localhost:3000") {
		t.Error("expected https://localhost:3000 to be rejected (scheme mismatch)")
	}
}

func TestWS_OriginWildcard(t *testing.T) {
	h := newTestWSHandler("*")

	if !h.isOriginAllowed("http://anything.com") {
		t.Error("expected any origin to be allowed with wildcard")
	}
	if !h.isOriginAllowed("") {
		t.Error("expected empty origin to be allowed with wildcard")
	}
}

func TestWS_OriginEmpty(t *testing.T) {
	h := newTestWSHandler("http://localhost:3000")

	if h.isOriginAllowed("") {
		t.Error("expected empty origin to be rejected when not wildcard")
	}
}

func TestWS_OriginMultiple(t *testing.T) {
	h := newTestWSHandler("http://localhost:3000, https://staging.example.com, https://prod.example.com")

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"http://localhost:3000", true},
		{"https://staging.example.com", true},
		{"https://prod.example.com", true},
		{"http://other.com", false},
		{"", false},
	}

	for _, tt := range tests {
		result := h.isOriginAllowed(tt.origin)
		if result != tt.allowed {
			t.Errorf("origin %q: expected allowed=%v, got %v", tt.origin, tt.allowed, result)
		}
	}
}

func TestWS_OriginCaseInsensitive(t *testing.T) {
	h := newTestWSHandler("http://localhost:3000")

	if !h.isOriginAllowed("HTTP://LOCALHOST:3000") {
		t.Error("expected case-insensitive match")
	}
}

func TestWS_NewWS_ParsesOrigins(t *testing.T) {
	h := newTestWSHandler("http://a.com, http://b.com ,http://c.com")

	if len(h.allowedOrigins) != 3 {
		t.Errorf("expected 3 origins, got %d", len(h.allowedOrigins))
	}

	expected := []string{"http://a.com", "http://b.com", "http://c.com"}
	for i, exp := range expected {
		if h.allowedOrigins[i] != exp {
			t.Errorf("origin %d: expected %q, got %q", i, exp, h.allowedOrigins[i])
		}
	}
}

func TestWS_OriginEmptyConfigDefaultsToWildcard(t *testing.T) {
	h := newTestWSHandler("")
	if len(h.allowedOrigins) != 1 || h.allowedOrigins[0] != "*" {
		t.Errorf("expected wildcard default, got %v", h.allowedOrigins)
	}
	if !h.isOriginAllowed("http://anything.com") {
		t.Error("empty config should default to wildcard (allow all)")
	}
}

func TestWS_OriginWhitespaceConfigDefaultsToWildcard(t *testing.T) {
	h := newTestWSHandler("  ,  , ")
	if len(h.allowedOrigins) != 1 || h.allowedOrigins[0] != "*" {
		t.Errorf("expected wildcard default, got %v", h.allowedOrigins)
	}
}

func TestWS_NewWS_Config(t *testing.T) {
	h := newTestWSHandler("*")

	if string(h.apiKey) != "test-api-key" {
		t.Errorf("expected apiKey 'test-api-key', got %q", string(h.apiKey))
	}
	if h.authTimeout != 5*time.Second {
		t.Errorf("expected authTimeout 5s, got %v", h.authTimeout)
	}
}

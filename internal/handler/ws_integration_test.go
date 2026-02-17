package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	fws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
)

type authResponse struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// startTestApp starts a fiber app with the WS handler and returns the ws URL
func startTestApp(t *testing.T, apiKey string, authTimeout time.Duration, allowedOrigins string) (string, *ws.Hub) {
	t.Helper()

	logger := zerolog.New(io.Discard)
	hub := ws.NewHub(logger)
	go hub.Run()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	wsHandler := NewWS(hub, apiKey, authTimeout, allowedOrigins, logger)
	app.Get("/ws", wsHandler.Upgrade)

	// Get a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	go func() {
		_ = app.Listen(addr)
	}()

	// Wait for server to start
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Cleanup(func() {
		hub.Shutdown()
		_ = app.Shutdown()
	})

	return fmt.Sprintf("ws://%s/ws", addr), hub
}

func TestIntegration_AuthValid(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Send auth
	auth := map[string]string{"type": "auth", "apiKey": "test-api-key"}
	if err := conn.WriteJSON(auth); err != nil {
		t.Fatalf("write auth failed: %v", err)
	}

	// Read response
	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected auth success, got failure: %s", resp.Message)
	}
	if resp.Type != "auth" {
		t.Errorf("expected type 'auth', got %q", resp.Type)
	}
}

func TestIntegration_AuthInvalidKey(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	auth := map[string]string{"type": "auth", "apiKey": "wrong-key"}
	if err := conn.WriteJSON(auth); err != nil {
		t.Fatalf("write auth failed: %v", err)
	}

	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if resp.Success {
		t.Error("expected auth failure")
	}
	if resp.Message != "Invalid API key" {
		t.Errorf("expected 'Invalid API key', got %q", resp.Message)
	}
}

func TestIntegration_AuthMalformedJSON(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.WriteMessage(fws.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if resp.Success {
		t.Error("expected auth failure")
	}
	if resp.Message != "Invalid message format" {
		t.Errorf("expected 'Invalid message format', got %q", resp.Message)
	}
}

func TestIntegration_AuthWrongType(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	msg := map[string]string{"type": "subscribe", "apiKey": "test-api-key"}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if resp.Success {
		t.Error("expected auth failure")
	}
	if resp.Message != "First message must be auth" {
		t.Errorf("expected 'First message must be auth', got %q", resp.Message)
	}
}

func TestIntegration_AuthEmptyKey(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	msg := map[string]string{"type": "auth", "apiKey": ""}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if resp.Success {
		t.Error("expected auth failure")
	}
	if resp.Message != "API key is required" {
		t.Errorf("expected 'API key is required', got %q", resp.Message)
	}
}

func TestIntegration_AuthTimeout(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 1*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Don't send auth, wait for timeout
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if resp.Success {
		t.Error("expected auth failure from timeout")
	}
	if resp.Message != "Authentication timeout" {
		t.Errorf("expected 'Authentication timeout', got %q", resp.Message)
	}
}

func TestIntegration_BroadcastAfterAuth(t *testing.T) {
	url, hub := startTestApp(t, "test-api-key", 5*time.Second, "*")

	conn, _, err := fws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Auth
	auth := map[string]string{"type": "auth", "apiKey": "test-api-key"}
	if err := conn.WriteJSON(auth); err != nil {
		t.Fatalf("write auth failed: %v", err)
	}

	var resp authResponse
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read auth response failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("auth should succeed: %s", resp.Message)
	}

	// Give hub time to process
	time.Sleep(100 * time.Millisecond)

	// Send a broadcast
	hub.Broadcast("60123456789", "qrcode", "test-qr-data")

	// Read broadcast
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read broadcast failed: %v", err)
	}

	var broadcastMsg map[string]interface{}
	if err := json.Unmarshal(msg, &broadcastMsg); err != nil {
		t.Fatalf("unmarshal broadcast failed: %v", err)
	}

	if broadcastMsg["event"] != "qrcode" {
		t.Errorf("expected event 'qrcode', got %v", broadcastMsg["event"])
	}
}

func TestIntegration_OriginRejected(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "http://localhost:3000")

	headers := http.Header{}
	headers.Set("Origin", "http://evil.com")

	_, resp, err := fws.DefaultDialer.Dial(url, headers)
	if err == nil {
		t.Fatal("expected dial to fail with rejected origin")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestIntegration_OriginAllowed(t *testing.T) {
	url, _ := startTestApp(t, "test-api-key", 5*time.Second, "http://localhost:3000")

	headers := http.Header{}
	headers.Set("Origin", "http://localhost:3000")

	conn, _, err := fws.DefaultDialer.Dial(url, headers)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	_ = conn.Close()
}

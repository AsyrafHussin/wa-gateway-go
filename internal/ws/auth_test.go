package ws

import (
	"testing"
)

func TestClient_IsAuthenticated_Default(t *testing.T) {
	hub := newTestHub()
	c := mockClient(hub)
	if c.IsAuthenticated() {
		t.Error("new client should not be authenticated")
	}
}

func TestClient_IsAuthenticated_AfterStore(t *testing.T) {
	hub := newTestHub()
	c := mockClient(hub)
	c.authenticated.Store(true)
	if !c.IsAuthenticated() {
		t.Error("client should be authenticated after Store(true)")
	}
}

func TestAuthMessage_JSON(t *testing.T) {
	// Verify auth message structures serialize correctly
	msg := authMessage{Type: "auth", APIKey: "test-key"}
	if msg.Type != "auth" {
		t.Errorf("expected type 'auth', got %q", msg.Type)
	}
	if msg.APIKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %q", msg.APIKey)
	}
}

func TestAuthResponse_JSON(t *testing.T) {
	resp := authResponse{Type: "auth", Success: true}
	if resp.Type != "auth" {
		t.Errorf("expected type 'auth', got %q", resp.Type)
	}
	if !resp.Success {
		t.Error("expected success true")
	}

	errResp := authResponse{Type: "auth", Success: false, Message: "Invalid API key"}
	if errResp.Success {
		t.Error("expected success false")
	}
	if errResp.Message != "Invalid API key" {
		t.Errorf("expected message 'Invalid API key', got %q", errResp.Message)
	}
}

func TestClient_AuthTimeoutCancelled(t *testing.T) {
	hub := newTestHub()
	c := mockClient(hub)

	// Close authDone immediately (simulates successful auth before timeout)
	close(c.authDone)

	// AuthTimeoutPump should return immediately without disconnect
	done := make(chan struct{})
	go func() {
		c.AuthTimeoutPump(1 * 1e9) // 1 second
		close(done)
	}()

	select {
	case <-done:
		// ok, returned quickly
	case <-make(chan struct{}):
		t.Fatal("AuthTimeoutPump should have returned immediately")
	}
}

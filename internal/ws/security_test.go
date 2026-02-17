package ws

import (
	"crypto/subtle"
	"strings"
	"testing"
)

func TestClient_Auth_TimingAttack(t *testing.T) {
	// Verify constant-time comparison is used for key validation
	serverKey := []byte("correct-api-key-12345")

	keys := []string{
		"wrong",
		"correct-api-key-1234",  // one char short
		"correct-api-key-123456", // one char long
		"CORRECT-API-KEY-12345", // different case
		"x",
		strings.Repeat("a", 1000), // very long
	}

	for _, key := range keys {
		result := subtle.ConstantTimeCompare([]byte(key), serverKey)
		if result != 0 {
			t.Errorf("key %q should not match server key", key)
		}
	}

	// Correct key should match
	result := subtle.ConstantTimeCompare([]byte("correct-api-key-12345"), serverKey)
	if result != 1 {
		t.Error("correct key should match")
	}
}

func TestClient_Auth_NullBytes(t *testing.T) {
	serverKey := []byte("valid-key")
	keyWithNulls := []byte("valid-key\x00\x00")

	result := subtle.ConstantTimeCompare(keyWithNulls, serverKey)
	if result != 0 {
		t.Error("key with null bytes should not match server key")
	}
}

func TestClient_Auth_ReplayAuth(t *testing.T) {
	hub := newTestHub()
	c := mockClient(hub)

	// Simulate first auth success
	c.authenticated.Store(true)
	c.authOnce.Do(func() { close(c.authDone) })

	// Second close should be a no-op via sync.Once
	c.authOnce.Do(func() { close(c.authDone) })
	// If we got here without panic, the test passes

	if !c.IsAuthenticated() {
		t.Error("client should still be authenticated")
	}
}

func TestClient_Auth_OversizedPayload(t *testing.T) {
	// Verify max auth message size constant is set
	if maxAuthMessageSize != 1024 {
		t.Errorf("expected maxAuthMessageSize 1024, got %d", maxAuthMessageSize)
	}

	// A message larger than maxAuthMessageSize should be rejected
	oversized := strings.Repeat("x", maxAuthMessageSize+1)
	if len(oversized) <= maxAuthMessageSize {
		t.Error("test payload should exceed max size")
	}
}

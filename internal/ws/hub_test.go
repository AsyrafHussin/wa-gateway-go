package ws

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

var testLogger = zerolog.New(io.Discard)

func newTestHub() *Hub {
	return NewHub(testLogger)
}

// mockClient creates a client with only a send channel (no real websocket)
func mockClient(hub *Hub) *Client {
	c := &Client{
		hub:      hub,
		conn:     nil,
		send:     make(chan []byte, 256),
		authDone: make(chan struct{}),
		apiKey:   []byte("test-key"),
		logger:   testLogger,
	}
	return c
}

// mockAuthenticatedClient creates a client that is already authenticated
func mockAuthenticatedClient(hub *Hub) *Client {
	c := mockClient(hub)
	c.authenticated.Store(true)
	return c
}

func TestNewHub(t *testing.T) {
	hub := newTestHub()
	if hub == nil {
		t.Fatal("expected non-nil Hub")
	}
	if hub.clients == nil {
		t.Fatal("expected non-nil clients map")
	}
	if hub.broadcast == nil {
		t.Fatal("expected non-nil broadcast channel")
	}
}

func TestHub_ClientCount_Empty(t *testing.T) {
	hub := newTestHub()
	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	client := mockClient(hub)

	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", hub.ClientCount())
	}

	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients after unregister, got %d", hub.ClientCount())
	}
}

func TestHub_RegisterMultipleClients(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	clients := make([]*Client, 5)
	for i := range clients {
		clients[i] = mockClient(hub)
		hub.Register(clients[i])
	}
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 5 {
		t.Errorf("expected 5 clients, got %d", hub.ClientCount())
	}

	// Unregister 2
	hub.Unregister(clients[0])
	hub.Unregister(clients[1])
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 3 {
		t.Errorf("expected 3 clients after unregister, got %d", hub.ClientCount())
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	client := mockAuthenticatedClient(hub)
	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast("60123456789", "qrcode", "qr-data-here")
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.send:
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if m.Event != "qrcode" {
			t.Errorf("expected event 'qrcode', got %q", m.Event)
		}
		if m.Token != "60123456789" {
			t.Errorf("expected token '60123456789', got %q", m.Token)
		}
	default:
		t.Error("expected message on client send channel")
	}
}

func TestHub_BroadcastWithMessage(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	client := mockAuthenticatedClient(hub)
	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	hub.BroadcastWithMessage("token123", "connection-error", "Disconnected")
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.send:
		var m Message
		if err := json.Unmarshal(msg, &m); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if m.Event != "connection-error" {
			t.Errorf("expected event 'connection-error', got %q", m.Event)
		}
		if m.Message != "Disconnected" {
			t.Errorf("expected message 'Disconnected', got %q", m.Message)
		}
		if m.Token != "token123" {
			t.Errorf("expected token 'token123', got %q", m.Token)
		}
	default:
		t.Error("expected message on client send channel")
	}
}

func TestHub_BroadcastToMultipleClients(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	clients := make([]*Client, 3)
	for i := range clients {
		clients[i] = mockAuthenticatedClient(hub)
		hub.Register(clients[i])
	}
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast("token", "test-event", "test-data")
	time.Sleep(50 * time.Millisecond)

	for i, client := range clients {
		select {
		case msg := <-client.send:
			var m Message
			if err := json.Unmarshal(msg, &m); err != nil {
				t.Fatalf("client %d: failed to unmarshal: %v", i, err)
			}
			if m.Event != "test-event" {
				t.Errorf("client %d: expected event 'test-event', got %q", i, m.Event)
			}
		default:
			t.Errorf("client %d: expected message", i)
		}
	}
}

func TestHub_BroadcastNoClients(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	// Should not panic with no clients
	hub.Broadcast("token", "event", "data")
	time.Sleep(50 * time.Millisecond)
}

func TestHub_Shutdown(t *testing.T) {
	hub := newTestHub()
	go hub.Run()

	client1 := mockClient(hub)
	client2 := mockClient(hub)
	hub.Register(client1)
	hub.Register(client2)
	time.Sleep(50 * time.Millisecond)

	hub.Shutdown()
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients after shutdown, got %d", hub.ClientCount())
	}
}

func TestHub_UnregisterNonExistentClient(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	client := mockClient(hub)
	// Unregister without registering — should not panic
	hub.Unregister(client)
	time.Sleep(50 * time.Millisecond)
}

func TestMessage_JSONSerialization(t *testing.T) {
	msg := Message{
		Event:   "connection-success",
		Token:   "60123456789",
		Data:    map[string]string{"status": "ok"},
		Message: "",
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Event != msg.Event {
		t.Errorf("expected event %q, got %q", msg.Event, decoded.Event)
	}
	if decoded.Token != msg.Token {
		t.Errorf("expected token %q, got %q", msg.Token, decoded.Token)
	}
}

func TestMessage_OmitsEmptyFields(t *testing.T) {
	msg := Message{
		Event: "test",
		Token: "123",
	}

	bytes, _ := json.Marshal(msg)
	str := string(bytes)

	// Data and Message should be omitted when empty
	if containsField(str, "data") {
		t.Error("expected 'data' to be omitted when nil")
	}
	if containsField(str, "message") {
		t.Error("expected 'message' to be omitted when empty")
	}
}

func containsField(jsonStr, field string) bool {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(jsonStr), &m)
	_, ok := m[field]
	return ok
}

func TestHub_BroadcastDataTypes(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	client := mockAuthenticatedClient(hub)
	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	tests := []struct {
		name  string
		event string
		data  interface{}
	}{
		{"string data", "event1", "simple-string"},
		{"map data", "event2", map[string]string{"key": "value"}},
		{"nil data", "event3", nil},
		{"bool data", "event4", true},
		{"int data", "event5", 42},
		{"slice data", "event6", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub.Broadcast("token", tt.event, tt.data)
			time.Sleep(50 * time.Millisecond)

			select {
			case msg := <-client.send:
				var m Message
				if err := json.Unmarshal(msg, &m); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if m.Event != tt.event {
					t.Errorf("expected event %q, got %q", tt.event, m.Event)
				}
			default:
				t.Error("expected message")
			}
		})
	}
}

func TestHub_BroadcastSkipsUnauthenticated(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	authed1 := mockAuthenticatedClient(hub)
	authed2 := mockAuthenticatedClient(hub)
	unauthed := mockClient(hub)

	hub.Register(authed1)
	hub.Register(authed2)
	hub.Register(unauthed)
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast("token", "test-event", "data")
	time.Sleep(50 * time.Millisecond)

	// Authenticated clients should receive
	for i, client := range []*Client{authed1, authed2} {
		select {
		case <-client.send:
			// ok
		default:
			t.Errorf("authed client %d: expected message", i)
		}
	}

	// Unauthenticated client should NOT receive
	select {
	case <-unauthed.send:
		t.Error("unauthenticated client should not receive broadcast")
	default:
		// ok
	}
}

func TestHub_BroadcastNoneAuthenticated(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Shutdown()

	clients := make([]*Client, 3)
	for i := range clients {
		clients[i] = mockClient(hub) // all unauthenticated
		hub.Register(clients[i])
	}
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast("token", "test-event", "data")
	time.Sleep(50 * time.Millisecond)

	for i, client := range clients {
		select {
		case <-client.send:
			t.Errorf("client %d: should not receive broadcast when unauthenticated", i)
		default:
			// ok
		}
	}
}

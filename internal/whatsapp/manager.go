package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	"github.com/AsyrafHussin/wa-gateway-go/config"
	"github.com/AsyrafHussin/wa-gateway-go/internal/webhook"
	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"
)

type SessionInfo struct {
	Token  string `json:"token"`
	Status string `json:"status"`
}

type DeviceManager struct {
	mu       sync.RWMutex
	sessions map[string]*DeviceSession
	config   *config.Config
	hub      *ws.Hub
	webhook  *webhook.Dispatcher
	logger   zerolog.Logger
}

func NewDeviceManager(cfg *config.Config, hub *ws.Hub, dispatcher *webhook.Dispatcher, logger zerolog.Logger) *DeviceManager {
	return &DeviceManager{
		sessions: make(map[string]*DeviceSession),
		config:   cfg,
		hub:      hub,
		webhook:  dispatcher,
		logger:   logger.With().Str("component", "device_manager").Logger(),
	}
}

func (m *DeviceManager) Connect(ctx context.Context, token string, method string) error {
	m.mu.Lock()
	if existing, ok := m.sessions[token]; ok {
		if existing.GetStatus() == StatusConnected {
			m.mu.Unlock()
			return fmt.Errorf("device already connected")
		}
		// Clean up old session
		existing.Disconnect(ctx)
		delete(m.sessions, token)
	}

	session, err := NewDeviceSession(token, m.config, m.hub, m.webhook, m.logger)
	if err != nil {
		m.mu.Unlock()
		return fmt.Errorf("failed to create session: %w", err)
	}
	m.sessions[token] = session
	m.mu.Unlock()

	if err := session.Connect(ctx, method); err != nil {
		// Connect failed — remove session from map and clean up resources
		m.mu.Lock()
		delete(m.sessions, token)
		m.mu.Unlock()
		session.Disconnect(ctx)
		return err
	}
	return nil
}

func (m *DeviceManager) Disconnect(ctx context.Context, token string) error {
	m.mu.Lock()
	session, ok := m.sessions[token]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("device not found")
	}
	delete(m.sessions, token)
	m.mu.Unlock()

	session.Logout(ctx)
	return nil
}

func (m *DeviceManager) SendText(ctx context.Context, token, to, text string) (*SendResult, error) {
	session, ok := m.GetSession(token)
	if !ok {
		return nil, fmt.Errorf("device not found")
	}
	if session.GetStatus() != StatusConnected {
		return nil, fmt.Errorf("device not connected")
	}
	return session.SendText(ctx, to, text)
}

func (m *DeviceManager) ValidatePhone(ctx context.Context, token, phone string) ([]ValidateResult, error) {
	session, ok := m.GetSession(token)
	if !ok {
		return nil, fmt.Errorf("device not found")
	}
	if session.GetStatus() != StatusConnected {
		return nil, fmt.Errorf("device not connected")
	}
	return session.ValidatePhone(ctx, phone)
}

func (m *DeviceManager) GetSession(token string) (*DeviceSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[token]
	return s, ok
}

func (m *DeviceManager) ListSessions() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]SessionInfo, 0, len(m.sessions))
	for token, s := range m.sessions {
		infos = append(infos, SessionInfo{
			Token:  token,
			Status: s.GetStatus().String(),
		})
	}
	return infos
}

func (m *DeviceManager) ShutdownAll(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for token, session := range m.sessions {
		m.logger.Info().Str("token", token).Msg("disconnecting device")
		session.Disconnect(ctx)
	}
	m.sessions = make(map[string]*DeviceSession)
}

// AutoReconnect scans the data directory and reconnects existing sessions.
func (m *DeviceManager) AutoReconnect(ctx context.Context) {
	sessionsDir := filepath.Join(m.config.DataDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		m.logger.Debug().Msg("no existing sessions to reconnect")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}
		token := strings.TrimSuffix(entry.Name(), ".db")
		m.logger.Info().Str("token", token).Msg("auto-reconnecting existing session")

		if err := m.Connect(ctx, token, "reconnect"); err != nil {
			m.logger.Error().Err(err).Str("token", token).Msg("failed to auto-reconnect")
		}
	}
}

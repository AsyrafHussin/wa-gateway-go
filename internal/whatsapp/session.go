package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/AsyrafHussin/wa-gateway-go/config"
	"github.com/AsyrafHussin/wa-gateway-go/internal/contacts"
	"github.com/AsyrafHussin/wa-gateway-go/internal/webhook"
	"github.com/AsyrafHussin/wa-gateway-go/internal/ws"

	_ "modernc.org/sqlite"
)

type SessionStatus int

const (
	StatusDisconnected SessionStatus = iota
	StatusConnecting
	StatusConnected
)

func (s SessionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	default:
		return "unknown"
	}
}

type DeviceSession struct {
	Token    string
	Client   *whatsmeow.Client
	device   *store.Device
	Contacts *contacts.Store
	status   SessionStatus
	mu       sync.RWMutex
	config   *config.Config
	hub      *ws.Hub
	webhook  *webhook.Dispatcher
	logger   zerolog.Logger
}

func NewDeviceSession(token string, cfg *config.Config, hub *ws.Hub, dispatcher *webhook.Dispatcher, logger zerolog.Logger) (*DeviceSession, error) {
	// Ensure directories exist
	sessionsDir := filepath.Join(cfg.DataDir, "sessions")
	contactsDir := filepath.Join(cfg.DataDir, "contacts")
	os.MkdirAll(sessionsDir, 0755)
	os.MkdirAll(contactsDir, 0755)

	// Open contact store
	contactStore, err := contacts.NewStore(filepath.Join(contactsDir, token+".db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create contact store: %w", err)
	}

	return &DeviceSession{
		Token:    token,
		Contacts: contactStore,
		status:   StatusDisconnected,
		config:   cfg,
		hub:      hub,
		webhook:  dispatcher,
		logger:   logger.With().Str("token", token).Logger(),
	}, nil
}

func (s *DeviceSession) Connect(ctx context.Context, method string) error {
	s.setStatus(StatusConnecting)

	dbPath := filepath.Join(s.config.DataDir, "sessions", s.Token+".db")
	container, err := sqlstore.New(ctx, "sqlite", "file:"+dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", waLog.Noop)
	if err != nil {
		s.setStatus(StatusDisconnected)
		return fmt.Errorf("failed to create store: %w", err)
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		s.setStatus(StatusDisconnected)
		return fmt.Errorf("failed to get device: %w", err)
	}
	s.device = device

	client := whatsmeow.NewClient(device, waLog.Noop)
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true
	s.Client = client

	client.AddEventHandler(s.handleEvent)

	// Reconnect existing session
	if method == "reconnect" && device.ID != nil {
		s.logger.Info().Msg("reconnecting existing session")
		return client.Connect()
	}

	// New device — QR or pairing code
	if device.ID == nil {
		switch method {
		case "code":
			if err := client.Connect(); err != nil {
				s.setStatus(StatusDisconnected)
				return fmt.Errorf("failed to connect: %w", err)
			}
			code, err := client.PairPhone(ctx, "+"+s.Token, true, whatsmeow.PairClientChrome, "wa-gateway-go")
			if err != nil {
				s.setStatus(StatusDisconnected)
				return fmt.Errorf("failed to get pairing code: %w", err)
			}
			s.hub.Broadcast(s.Token, "pairing-code", map[string]string{"code": code})
			return nil

		default: // "qr"
			qrChan, err := client.GetQRChannel(ctx)
			if err != nil {
				s.setStatus(StatusDisconnected)
				return fmt.Errorf("failed to get QR channel: %w", err)
			}

			if err := client.Connect(); err != nil {
				s.setStatus(StatusDisconnected)
				return fmt.Errorf("failed to connect: %w", err)
			}

			// Process QR events in background
			go s.processQRChannel(qrChan)
			return nil
		}
	}

	// Existing session, just connect
	return client.Connect()
}

func (s *DeviceSession) processQRChannel(qrChan <-chan whatsmeow.QRChannelItem) {
	for evt := range qrChan {
		switch evt.Event {
		case "code":
			s.logger.Debug().Msg("QR code received")
			s.hub.Broadcast(s.Token, "qrcode", evt.Code)
		case "success":
			s.logger.Info().Msg("QR pairing successful")
		case "timeout":
			s.logger.Warn().Msg("QR code timeout")
			s.hub.BroadcastWithMessage(s.Token, "connection-error", "QR code expired")
			s.setStatus(StatusDisconnected)
		}
	}
}

func (s *DeviceSession) Disconnect(ctx context.Context) {
	if s.Client != nil {
		s.Client.Disconnect()
	}
	if s.Contacts != nil {
		s.Contacts.Close()
	}
	s.setStatus(StatusDisconnected)
}

func (s *DeviceSession) Logout(ctx context.Context) {
	if s.Client != nil {
		s.Client.Logout(ctx)
	}
	s.Disconnect(ctx)

	// Remove session DB
	dbPath := filepath.Join(s.config.DataDir, "sessions", s.Token+".db")
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")
}

func (s *DeviceSession) GetStatus() SessionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *DeviceSession) setStatus(status SessionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

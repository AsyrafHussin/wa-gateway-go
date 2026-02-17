package whatsapp

import (
	"context"
	"strings"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (s *DeviceSession) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		s.setStatus(StatusConnected)
		s.logger.Info().Msg("connected to WhatsApp")
		s.hub.Broadcast(s.Token, "connection-success", nil)
		s.webhook.Send("device.connected", s.Token, nil)
		if err := s.Client.SendPresence(context.Background(), types.PresenceAvailable); err != nil {
			s.logger.Error().Err(err).Msg("failed to send presence")
		}

	case *events.Disconnected:
		s.setStatus(StatusDisconnected)
		s.logger.Warn().Msg("disconnected from WhatsApp")
		s.hub.BroadcastWithMessage(s.Token, "connection-error", "Disconnected")
		s.webhook.Send("device.disconnected", s.Token, nil)

	case *events.LoggedOut:
		s.setStatus(StatusDisconnected)
		s.logger.Warn().Msg("logged out from WhatsApp")
		s.hub.BroadcastWithMessage(s.Token, "connection-error", "Logged out")
		s.webhook.Send("device.logged_out", s.Token, nil)

	case *events.Message:
		if !v.Info.IsFromMe {
			s.captureContact(v.Info.Sender.User, v.Info.PushName)
		}

	case *events.Receipt:
		s.webhook.Send("message.receipt", s.Token, map[string]interface{}{
			"type":       string(v.Type),
			"messageIds": v.MessageIDs,
			"from":       v.Sender.String(),
			"timestamp":  v.Timestamp,
		})

	case *events.HistorySync:
		go s.processHistorySync(v)

	case *events.StreamReplaced:
		s.setStatus(StatusDisconnected)
		s.logger.Warn().Msg("stream replaced by another client")

	case *events.ConnectFailure:
		s.setStatus(StatusDisconnected)
		s.logger.Error().Str("reason", v.Reason.String()).Msg("connection failure")
		s.hub.BroadcastWithMessage(s.Token, "connection-error", "Connection failed: "+v.Reason.String())

	case *events.TemporaryBan:
		s.logger.Error().Int("code", int(v.Code)).Msg("temporary ban")

	case *events.PairSuccess:
		s.logger.Info().Msg("pairing successful")
	}
}

func (s *DeviceSession) captureContact(phone, pushName string) {
	if phone == "" {
		return
	}
	if err := s.Contacts.Upsert(phone, pushName, "message"); err != nil {
		s.logger.Error().Err(err).Str("phone", phone).Msg("failed to capture contact")
		return
	}
	s.webhook.Send("contacts.new", s.Token, map[string]string{
		"phone": phone,
		"name":  pushName,
	})
}

func (s *DeviceSession) processHistorySync(evt *events.HistorySync) {
	data := evt.Data
	var synced []map[string]string

	for _, conv := range data.GetConversations() {
		chatJID := conv.GetID()
		if strings.HasSuffix(chatJID, "@s.whatsapp.net") {
			phone := strings.TrimSuffix(chatJID, "@s.whatsapp.net")
			name := conv.GetDisplayName()
			if err := s.Contacts.Upsert(phone, name, "history"); err != nil {
				s.logger.Error().Err(err).Str("phone", phone).Msg("failed to capture history contact")
				continue
			}
			synced = append(synced, map[string]string{
				"phone": phone,
				"name":  name,
			})
		}
	}

	if len(synced) > 0 {
		s.logger.Info().Int("count", len(synced)).Msg("contacts captured from history sync")
		s.webhook.Send("contacts.sync", s.Token, synced)
	}
}

package whatsapp

import (
	"context"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type SendResult struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type ValidateResult struct {
	Phone        string `json:"phone"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	JID          string `json:"jid,omitempty"`
}

func (s *DeviceSession) SendText(ctx context.Context, to, text string) (*SendResult, error) {
	jid := types.NewJID(to, types.DefaultUserServer)

	// Typing indicator
	_ = s.Client.SendChatPresence(ctx, jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)

	// Simulate typing delay (respects context cancellation)
	delay := time.Duration(s.config.TypingDelay) * time.Millisecond
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	// Stop typing
	_ = s.Client.SendChatPresence(ctx, jid, types.ChatPresencePaused, types.ChatPresenceMediaText)

	// Send message
	resp, err := s.Client.SendMessage(ctx, jid, &waE2E.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		return nil, err
	}

	return &SendResult{
		ID:        resp.ID,
		Timestamp: resp.Timestamp,
	}, nil
}

func (s *DeviceSession) ValidatePhone(ctx context.Context, phone string) ([]ValidateResult, error) {
	// whatsmeow requires "+" prefix
	results, err := s.Client.IsOnWhatsApp(ctx, []string{"+" + phone})
	if err != nil {
		return nil, err
	}

	var validated []ValidateResult
	for _, r := range results {
		validated = append(validated, ValidateResult{
			Phone:        phone,
			IsOnWhatsApp: r.IsIn,
			JID:          r.JID.String(),
		})
	}

	return validated, nil
}

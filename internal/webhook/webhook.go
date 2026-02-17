package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Dispatcher struct {
	url     string
	secret  string
	timeout time.Duration
	client  *http.Client
	logger  zerolog.Logger
}

type Payload struct {
	Event     string      `json:"event"`
	Token     string      `json:"token"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

func NewDispatcher(url, secret string, timeoutMs int, logger zerolog.Logger) *Dispatcher {
	return &Dispatcher{
		url:     url,
		secret:  secret,
		timeout: time.Duration(timeoutMs) * time.Millisecond,
		client: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
		logger: logger.With().Str("component", "webhook").Logger(),
	}
}

func (d *Dispatcher) Send(event, token string, data interface{}) {
	if d.url == "" {
		return
	}

	go d.dispatch(event, token, data)
}

func (d *Dispatcher) dispatch(event, token string, data interface{}) {
	payload := Payload{
		Event:     event,
		Token:     token,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		d.logger.Error().Err(err).Str("event", event).Msg("failed to marshal webhook payload")
		return
	}

	req, err := http.NewRequest("POST", d.url, bytes.NewReader(body))
	if err != nil {
		d.logger.Error().Err(err).Str("event", event).Msg("failed to create webhook request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "wa-gateway-go/1.0")

	if d.secret != "" {
		mac := hmac.New(sha256.New, []byte(d.secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature", sig)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		d.logger.Error().Err(err).Str("event", event).Msg("webhook request failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		d.logger.Warn().Int("status", resp.StatusCode).Str("event", event).Msg("webhook returned error status")
	} else {
		d.logger.Debug().Str("event", event).Msg("webhook delivered")
	}
}

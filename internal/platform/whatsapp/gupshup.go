package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GupshupProvider sends WhatsApp messages via Gupshup.
type GupshupProvider struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NewGupshupProvider constructs a Gupshup provider.
func NewGupshupProvider(apiKey string) *GupshupProvider {
	return &GupshupProvider{
		APIKey:  apiKey,
		BaseURL: "https://api.gupshup.io/wa/api/v1",
		Client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// SendText sends a plain text message.
func (p *GupshupProvider) SendText(ctx context.Context, phone, message string) (*SendResult, error) {
	form := url.Values{}
	form.Set("channel", "whatsapp")
	form.Set("source", "")
	form.Set("destination", strings.TrimPrefix(phone, "+"))
	form.Set("message", `{"type":"text","text":"`+escapeJSON(message)+`"}`)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/msg", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", p.APIKey)
	return p.do(req)
}

// SendTemplate sends a pre-approved template message.
func (p *GupshupProvider) SendTemplate(ctx context.Context, phone, templateName string, params map[string]string) (*SendResult, error) {
	payload := map[string]any{
		"type":     "template",
		"template": map[string]any{"id": templateName, "params": params},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("channel", "whatsapp")
	form.Set("source", "")
	form.Set("destination", strings.TrimPrefix(phone, "+"))
	form.Set("message", string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/msg", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", p.APIKey)
	return p.do(req)
}

func (p *GupshupProvider) do(req *http.Request) (*SendResult, error) {
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gupshup status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		MessageID string `json:"messageId"`
	}
	_ = json.Unmarshal(bytes.TrimSpace(body), &out)
	return &SendResult{ProviderRef: out.MessageID}, nil
}

func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return strings.Trim(string(b), `"`)
}

package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MSG91Provider sends SMS through the MSG91 flow API.
type MSG91Provider struct {
	AuthKey string
	Sender  string
	BaseURL string
	Client  *http.Client
}

// NewMSG91Provider constructs an MSG91 provider.
func NewMSG91Provider(authKey, sender string) *MSG91Provider {
	return &MSG91Provider{
		AuthKey: authKey,
		Sender:  sender,
		BaseURL: "https://control.msg91.com/api/v5",
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

type msg91Request struct {
	Sender    string `json:"sender"`
	Route     string `json:"route"`
	Country   string `json:"country"`
	Template  string `json:"DLT_TE_ID,omitempty"`
	Recipient []struct {
		Mobile  string `json:"mobiles"`
		Message string `json:"message"`
	} `json:"recipients"`
}

// Send dispatches a single SMS via MSG91.
func (p *MSG91Provider) Send(ctx context.Context, phone, message, templateID string) (*SendResult, error) {
	req := msg91Request{Sender: p.Sender, Route: "4", Country: "91", Template: templateID}
	req.Recipient = append(req.Recipient, struct {
		Mobile  string `json:"mobiles"`
		Message string `json:"message"`
	}{Mobile: phone, Message: message})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/flow/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authkey", p.AuthKey)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("msg91 returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return &SendResult{ProviderRef: "msg91"}, nil
}

package sms

import (
	"context"
	"log/slog"
)

// ConsoleProvider logs messages instead of sending them. It is used in
// development and as a safe fallback when no provider is configured.
type ConsoleProvider struct {
	Logger *slog.Logger
}

// NewConsoleProvider creates a console (dev) SMS provider.
func NewConsoleProvider(logger *slog.Logger) *ConsoleProvider {
	return &ConsoleProvider{Logger: logger}
}

// Send logs the message at debug level (OTP bodies are not logged in production).
func (p *ConsoleProvider) Send(ctx context.Context, phone, message, templateID string) (*SendResult, error) {
	p.Logger.Info("sms_console_send", "phone_last4", last4(phone), "template", templateID)
	p.Logger.Debug("sms_console_body", "message", message)
	return &SendResult{ProviderRef: "console"}, nil
}

func last4(s string) string {
	if len(s) <= 4 {
		return s
	}
	return s[len(s)-4:]
}

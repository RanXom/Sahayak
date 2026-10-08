package whatsapp

import (
	"context"
	"log/slog"
)

// ConsoleProvider logs messages instead of sending them (development/fallback).
type ConsoleProvider struct {
	Logger *slog.Logger
}

// NewConsoleProvider constructs a console WhatsApp provider.
func NewConsoleProvider(logger *slog.Logger) *ConsoleProvider {
	return &ConsoleProvider{Logger: logger}
}

// SendTemplate logs a template message.
func (p *ConsoleProvider) SendTemplate(_ context.Context, phone, templateName string, _ map[string]string) (*SendResult, error) {
	p.Logger.Info("whatsapp_console_template", "phone_last4", last4(phone), "template", templateName)
	return &SendResult{ProviderRef: "console"}, nil
}

// SendText logs a text message.
func (p *ConsoleProvider) SendText(_ context.Context, phone, message string) (*SendResult, error) {
	p.Logger.Info("whatsapp_console_text", "phone_last4", last4(phone))
	p.Logger.Debug("whatsapp_console_body", "message", message)
	return &SendResult{ProviderRef: "console"}, nil
}

func last4(s string) string {
	if len(s) <= 4 {
		return s
	}
	return s[len(s)-4:]
}

// Package whatsapp defines the WhatsApp Business provider abstraction.
package whatsapp

import "context"

// SendResult is returned after a message is accepted by the provider.
type SendResult struct {
	ProviderRef string
}

// Provider sends WhatsApp messages.
type Provider interface {
	SendTemplate(ctx context.Context, phone, templateName string, params map[string]string) (*SendResult, error)
	SendText(ctx context.Context, phone, message string) (*SendResult, error)
}

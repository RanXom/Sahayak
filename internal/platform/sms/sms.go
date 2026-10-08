// Package sms defines the SMS provider abstraction and implementations.
package sms

import "context"

// SendResult is returned by a provider after accepting a message.
type SendResult struct {
	ProviderRef string
}

// Provider sends transactional SMS messages.
type Provider interface {
	Send(ctx context.Context, phone, message, templateID string) (*SendResult, error)
}

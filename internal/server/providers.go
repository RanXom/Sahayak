package server

import (
	"github.com/sahayak/sahayak/internal/platform/sms"
	"github.com/sahayak/sahayak/internal/platform/whatsapp"
)

// smsProvider selects the configured SMS provider, falling back to a console
// provider when no credentials are configured (development).
func (s *Server) smsProvider() sms.Provider {
	switch s.cfg.SMSProvider {
	case "msg91":
		if s.cfg.MSG91AuthKey != "" {
			return sms.NewMSG91Provider(s.cfg.MSG91AuthKey, s.cfg.MSG91Sender)
		}
	}
	return sms.NewConsoleProvider(s.logger)
}

// whatsappProvider selects the configured WhatsApp provider, falling back to a
// console provider in development.
func (s *Server) whatsappProvider() whatsapp.Provider {
	switch s.cfg.WhatsAppProvider {
	case "gupshup":
		if s.cfg.WhatsAppAPIKey != "" {
			return whatsapp.NewGupshupProvider(s.cfg.WhatsAppAPIKey)
		}
	}
	return whatsapp.NewConsoleProvider(s.logger)
}

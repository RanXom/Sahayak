package notification

import (
	"context"
	"errors"
	"time"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/crypto"
	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/platform/sms"
	"github.com/sahayak/sahayak/internal/platform/whatsapp"
)

// Enqueuer enqueues background notification jobs.
type Enqueuer interface {
	InsertNotification(ctx context.Context, id string) error
}

// PhoneProvider returns a user's encrypted phone number.
type PhoneProvider interface {
	EncryptedPhone(ctx context.Context, userID string) (string, error)
}

// Service implements notification use cases.
type Service struct {
	repo     *Repository
	phones   PhoneProvider
	cipher   *crypto.Cipher
	whatsapp whatsapp.Provider
	sms      sms.Provider
	enqueuer Enqueuer
	audit    audit.Logger
}

// Deps are the dependencies for the notification service.
type Deps struct {
	Repo     *Repository
	Phones   PhoneProvider
	Cipher   *crypto.Cipher
	WhatsApp whatsapp.Provider
	SMS      sms.Provider
	Enqueuer Enqueuer
	Audit    audit.Logger
}

// NewService constructs the notification service.
func NewService(d Deps) *Service {
	return &Service{
		repo:     d.Repo,
		phones:   d.Phones,
		cipher:   d.Cipher,
		whatsapp: d.WhatsApp,
		sms:      d.SMS,
		enqueuer: d.Enqueuer,
		audit:    d.Audit,
	}
}

// SetEnqueuer late-binds the job enqueuer (avoids an initialization cycle).
func (s *Service) SetEnqueuer(e Enqueuer) { s.enqueuer = e }

// GetPreferences returns a user's preferences.
func (s *Service) GetPreferences(ctx context.Context, userID string) ([]Preference, error) {
	out, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, domain.Internal("Could not load notification preferences", err)
	}
	return out, nil
}

// UpdatePreferences updates a channel preference.
func (s *Service) UpdatePreferences(ctx context.Context, userID, channel string, enabled bool) ([]Preference, error) {
	switch channel {
	case ChannelWhatsApp, ChannelSMS:
	default:
		return nil, domain.InvalidInput("Invalid channel")
	}
	if err := s.repo.UpsertPreference(ctx, userID, channel, enabled); err != nil {
		return nil, domain.Internal("Could not update preferences", err)
	}
	return s.repo.GetPreferences(ctx, userID)
}

// List returns a user's notifications.
func (s *Service) List(ctx context.Context, userID string, limit int) ([]Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	out, err := s.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, domain.Internal("Could not list notifications", err)
	}
	return out, nil
}

// NotifyRequest describes a notification to create.
type NotifyRequest struct {
	UserID        string
	ApplicationID *string
	Channel       string
	TemplateKey   string
	Language      string
	Params        map[string]string
	ScheduledFor  *time.Time
}

// Notify creates a notification (respecting preferences) and enqueues dispatch.
func (s *Service) Notify(ctx context.Context, req NotifyRequest) (string, error) {
	channel := req.Channel
	if channel == "" {
		channel = ChannelWhatsApp
	}
	// Respect opt-in preferences.
	if enabled, err := s.repo.IsEnabled(ctx, req.UserID, channel); err == nil && !enabled {
		// Fall back to SMS if WhatsApp not enabled, otherwise skip.
		if channel == ChannelWhatsApp {
			if smsEnabled, _ := s.repo.IsEnabled(ctx, req.UserID, ChannelSMS); smsEnabled {
				channel = ChannelSMS
			} else {
				return "", nil
			}
		} else {
			return "", nil
		}
	}

	content := Render(req.TemplateKey, req.Language, req.Params)
	id, err := s.repo.Create(ctx, Notification{
		UserID:        req.UserID,
		ApplicationID: req.ApplicationID,
		Channel:       channel,
		TemplateKey:   req.TemplateKey,
		Content:       content,
		ScheduledFor:  req.ScheduledFor,
	})
	if err != nil {
		return "", domain.Internal("Could not create notification", err)
	}
	if req.ScheduledFor == nil && s.enqueuer != nil {
		if err := s.enqueuer.InsertNotification(ctx, id); err != nil {
			// Notification remains pending; a periodic sweep can pick it up.
			return id, nil
		}
	}
	return id, nil
}

// Dispatch sends a queued notification. It is idempotent: already-sent
// notifications are skipped. Returning an error triggers River's retry policy.
func (s *Service) Dispatch(ctx context.Context, id string) error {
	n, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if n.Status == StatusSent || n.Status == StatusDelivered {
		return nil
	}

	phone, err := s.destinationPhone(ctx, n.UserID)
	if err != nil || phone == "" {
		_ = s.repo.MarkFailed(ctx, id, "no phone number on file")
		return nil
	}

	ref, sendErr := s.send(ctx, n.Channel, phone, n.Content)
	if sendErr != nil {
		// Fall back from WhatsApp to SMS.
		if n.Channel == ChannelWhatsApp && s.sms != nil {
			if ref2, err2 := s.send(ctx, ChannelSMS, phone, n.Content); err2 == nil {
				_ = s.repo.MarkSent(ctx, id, ref2)
				return nil
			}
		}
		_ = s.repo.MarkFailed(ctx, id, sendErr.Error())
		return sendErr
	}
	if err := s.repo.MarkSent(ctx, id, ref); err != nil {
		return err
	}
	s.audit.Log(ctx, audit.Entry{
		Action:       "notification.send",
		ResourceType: "notification",
		ResourceID:   id,
		Details:      map[string]any{"channel": n.Channel, "template": n.TemplateKey, "status": "sent"},
	})
	return nil
}

func (s *Service) send(ctx context.Context, channel, phone, content string) (string, error) {
	switch channel {
	case ChannelSMS:
		if s.sms == nil {
			return "", errors.New("sms provider not configured")
		}
		res, err := s.sms.Send(ctx, phone, content, "")
		if err != nil {
			return "", err
		}
		return res.ProviderRef, nil
	default:
		if s.whatsapp == nil {
			return "", errors.New("whatsapp provider not configured")
		}
		res, err := s.whatsapp.SendText(ctx, phone, content)
		if err != nil {
			return "", err
		}
		return res.ProviderRef, nil
	}
}

func (s *Service) destinationPhone(ctx context.Context, userID string) (string, error) {
	if s.phones == nil || s.cipher == nil {
		return "", nil
	}
	enc, err := s.phones.EncryptedPhone(ctx, userID)
	if err != nil || enc == "" {
		return "", err
	}
	return s.cipher.Decrypt(enc)
}

// DispatchArgs is the worker handler for a notification job.
func (s *Service) DispatchArgs(ctx context.Context, id string) error {
	return s.Dispatch(ctx, id)
}

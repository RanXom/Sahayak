package auth

import (
	"context"
	"errors"
	"time"

	"github.com/sahayak/sahayak/internal/audit"
	"github.com/sahayak/sahayak/internal/crypto"
	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/platform/sms"
)

// OTPTTL is how long an OTP challenge remains valid.
const OTPTTL = 5 * time.Minute

// Service implements the authentication use cases.
type Service struct {
	repo        *Repository
	tokens      *TokenManager
	sms         sms.Provider
	audit       audit.Logger
	cipher      *crypto.Cipher
	pepper      string
	demoOTP     string
	otpPer10Min int
	otpPerDay   int
}

// Deps are the dependencies for the auth service.
type Deps struct {
	Repo        *Repository
	Tokens      *TokenManager
	SMS         sms.Provider
	Audit       audit.Logger
	Cipher      *crypto.Cipher
	Pepper      string
	DemoOTP     string
	OTPPer10Min int
	OTPPerDay   int
}

// NewService constructs the auth service.
func NewService(d Deps) *Service {
	return &Service{
		repo:        d.Repo,
		tokens:      d.Tokens,
		sms:         d.SMS,
		audit:       d.Audit,
		cipher:      d.Cipher,
		pepper:      d.Pepper,
		demoOTP:     d.DemoOTP,
		otpPer10Min: d.OTPPer10Min,
		otpPerDay:   d.OTPPerDay,
	}
}

// RequestOTPResult is returned after requesting an OTP.
type RequestOTPResult struct {
	ChallengeID string `json:"challenge_id"`
	ExpiresIn   int    `json:"expires_in"`
	// DemoOTP is populated only when demo mode is enabled, so the UI can show
	// the code without an SMS provider.
	DemoOTP string `json:"demo_otp,omitempty"`
}

// AuthResult is returned on successful authentication.
type AuthResult struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         domain.User `json:"user"`
}

// RequestOTP generates and sends an OTP to the given phone number.
func (s *Service) RequestOTP(ctx context.Context, rawPhone, ip, ua string) (*RequestOTPResult, error) {
	phone, err := NormalizePhone(rawPhone)
	if err != nil {
		return nil, domain.InvalidInput("Please provide a valid phone number")
	}
	phoneHash := HashPhone(phone, s.pepper)

	// Demo mode bypasses per-phone OTP rate limits so a shared venue IP cannot
	// lock out presenters or judges.
	if s.demoOTP == "" {
		if s.otpPer10Min > 0 {
			n, err := s.repo.CountRecentOTPRequests(ctx, phoneHash, time.Now().Add(-10*time.Minute))
			if err != nil {
				return nil, domain.Internal("Could not process OTP request", err)
			}
			if n >= s.otpPer10Min {
				return nil, domain.RateLimited("Too many OTP requests. Please try again in 10 minutes.")
			}
		}
		if s.otpPerDay > 0 {
			n, err := s.repo.CountRecentOTPRequests(ctx, phoneHash, time.Now().Add(-24*time.Hour))
			if err != nil {
				return nil, domain.Internal("Could not process OTP request", err)
			}
			if n >= s.otpPerDay {
				return nil, domain.RateLimited("Daily OTP limit reached. Please try again tomorrow.")
			}
		}
	}

	otp, err := GenerateOTP()
	if err != nil {
		return nil, domain.Internal("Could not generate OTP", err)
	}
	otpHash, err := HashOTP(otp)
	if err != nil {
		return nil, domain.Internal("Could not secure OTP", err)
	}

	encPhone := ""
	if s.cipher != nil {
		if enc, err := s.cipher.Encrypt(phone); err == nil {
			encPhone = enc
		}
	}

	challengeID, err := s.repo.CreateOTPChallenge(ctx, phoneHash, Last4(phone), encPhone, otpHash, "login", time.Now().Add(OTPTTL))
	if err != nil {
		return nil, domain.Internal("Could not create OTP challenge", err)
	}

	// Demo mode: skip SMS and return the fixed code to the client.
	if s.demoOTP != "" {
		s.audit.Log(ctx, audit.Entry{Action: "auth.otp.request", ResourceType: "otp", ResourceID: challengeID, IPAddress: ip, UserAgent: ua, Details: map[string]any{"demo": true}})
		return &RequestOTPResult{ChallengeID: challengeID, ExpiresIn: int(OTPTTL.Seconds()), DemoOTP: s.demoOTP}, nil
	}

	message := "Your Sahayak verification code is " + otp + ". It is valid for 5 minutes. Do not share it with anyone."
	if _, err := s.sms.Send(ctx, phone, message, "login_otp"); err != nil {
		s.audit.Log(ctx, audit.Entry{Action: "auth.otp.send_failed", ResourceType: "otp", ResourceID: challengeID, IPAddress: ip, UserAgent: ua})
		return nil, domain.Unavailable("Could not send OTP right now. Please try again shortly.")
	}

	s.audit.Log(ctx, audit.Entry{Action: "auth.otp.request", ResourceType: "otp", ResourceID: challengeID, IPAddress: ip, UserAgent: ua})
	return &RequestOTPResult{ChallengeID: challengeID, ExpiresIn: int(OTPTTL.Seconds())}, nil
}

// VerifyOTP validates an OTP and returns tokens, creating the user if new.
func (s *Service) VerifyOTP(ctx context.Context, challengeID, otp, ip, ua string) (*AuthResult, error) {
	challenge, err := s.repo.GetOTPChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.InvalidInput("Invalid or expired verification code")
		}
		return nil, domain.Internal("Could not verify OTP", err)
	}
	if challenge.ConsumedAt != nil {
		return nil, domain.InvalidInput("This verification code has already been used")
	}
	if time.Now().After(challenge.ExpiresAt) {
		return nil, domain.InvalidInput("This verification code has expired")
	}
	if challenge.Attempts >= challenge.MaxAttempts {
		return nil, domain.RateLimited("Too many incorrect attempts. Please request a new code.")
	}

	valid := false
	if s.demoOTP != "" && otp == s.demoOTP {
		valid = true
	} else if VerifyOTP(challenge.OTPHash, otp) {
		valid = true
	}
	if !valid {
		_ = s.repo.IncrementOTPAttempts(ctx, challenge.ID)
		s.audit.Log(ctx, audit.Entry{Action: "auth.otp.failed", ResourceType: "otp", ResourceID: challenge.ID, IPAddress: ip, UserAgent: ua})
		return nil, domain.InvalidInput("Incorrect verification code")
	}

	if err := s.repo.ConsumeOTPChallenge(ctx, challenge.ID); err != nil {
		return nil, domain.Internal("Could not complete verification", err)
	}

	user, err := s.repo.FindUserByPhoneHash(ctx, challenge.PhoneHash)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, domain.Internal("Could not look up user", err)
		}
		user, err = s.repo.CreateUser(ctx, challenge.PhoneHash, challenge.PhoneLast4, challenge.PhoneEncrypted, "hi", domain.RoleCitizen)
		if err != nil {
			return nil, domain.Internal("Could not create user", err)
		}
	}

	if err := s.repo.TouchLastLogin(ctx, user.ID); err != nil {
		return nil, domain.Internal("Could not update login", err)
	}

	result, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}
	s.audit.Log(ctx, audit.Entry{
		ActorID:      &user.ID,
		ActorRole:    user.Role,
		Action:       "auth.login",
		ResourceType: "user",
		ResourceID:   user.ID,
		IPAddress:    ip,
		UserAgent:    ua,
	})
	return result, nil
}

// Refresh rotates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	hash := HashRefreshToken(refreshToken)
	userID, expiresAt, revokedAt, err := s.repo.GetRefreshToken(ctx, hash)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.Unauthorized("Invalid refresh token")
		}
		return nil, domain.Internal("Could not refresh token", err)
	}
	if revokedAt != nil || time.Now().After(expiresAt) {
		return nil, domain.Unauthorized("Refresh token expired or revoked")
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, domain.Unauthorized("User no longer exists")
	}

	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, domain.Internal("Could not rotate refresh token", err)
	}
	return s.issueTokens(ctx, user)
}

// Logout revokes a refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := HashRefreshToken(refreshToken)
	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return domain.Internal("Could not log out", err)
	}
	return nil
}

// Me returns the current user.
func (s *Service) Me(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, domain.NotFound("User not found")
		}
		return nil, domain.Internal("Could not load user", err)
	}
	return user, nil
}

func (s *Service) issueTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	access, err := s.tokens.IssueAccessToken(user.ID, user.Role, user.PreferredLang)
	if err != nil {
		return nil, domain.Internal("Could not issue access token", err)
	}
	refresh, refreshHash, err := NewRefreshToken()
	if err != nil {
		return nil, domain.Internal("Could not issue refresh token", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(s.tokens.RefreshTTL())); err != nil {
		return nil, domain.Internal("Could not store refresh token", err)
	}
	return &AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.tokens.AccessTTL().Seconds()),
		User:         *user,
	}, nil
}

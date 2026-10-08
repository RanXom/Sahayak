package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// TokenVerifier validates a Firebase ID token and returns the verified phone
// number (E.164) and Firebase UID.
type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (phone, uid string, err error)
}

// FirebaseVerifier is a TokenVerifier backed by the Firebase Admin SDK.
type FirebaseVerifier struct {
	client *auth.Client
}

// NewFirebaseVerifier constructs a verifier for the given project. credentials
// is the raw service-account JSON; when empty, Application Default Credentials
// are used (e.g. GOOGLE_APPLICATION_CREDENTIALS).
func NewFirebaseVerifier(ctx context.Context, projectID string, credentials []byte) (*FirebaseVerifier, error) {
	cfg := &firebase.Config{ProjectID: projectID}
	var opts []option.ClientOption
	if len(credentials) > 0 {
		opts = append(opts, option.WithCredentialsJSON(credentials))
	}
	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase auth client: %w", err)
	}
	return &FirebaseVerifier{client: client}, nil
}

// VerifyIDToken verifies the token signature, audience, issuer and expiry via
// Google's public keys and returns the phone_number claim.
func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, idToken string) (string, string, error) {
	if idToken == "" {
		return "", "", fmt.Errorf("empty id token")
	}
	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", "", fmt.Errorf("verify firebase id token: %w", err)
	}
	phone, _ := token.Claims["phone_number"].(string)
	if phone == "" {
		return "", "", fmt.Errorf("firebase token has no phone_number claim")
	}
	return phone, token.UID, nil
}

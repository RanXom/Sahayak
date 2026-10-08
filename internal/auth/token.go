package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sahayak/sahayak/internal/domain"
)

// TokenManager issues and verifies JWTs and opaque refresh tokens.
type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenManager constructs a TokenManager.
func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// AccessTTL returns the configured access-token lifetime.
func (m *TokenManager) AccessTTL() time.Duration { return m.accessTTL }

// RefreshTTL returns the configured refresh-token lifetime.
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }

type claims struct {
	Role string `json:"role"`
	Lang string `json:"lang"`
	jwt.RegisteredClaims
}

// IssueAccessToken signs a short-lived access token for the user.
func (m *TokenManager) IssueAccessToken(userID, role, lang string) (string, error) {
	now := time.Now()
	c := claims{
		Role: role,
		Lang: lang,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
			Issuer:    "sahayak",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(m.secret)
}

// VerifyAccessToken validates an access token and returns its claims.
func (m *TokenManager) VerifyAccessToken(tokenStr string) (*domain.Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer("sahayak"))
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return &domain.Claims{UserID: c.Subject, Role: c.Role, Lang: c.Lang}, nil
}

// GenerateOTP returns a cryptographically random 6-digit OTP.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashOTP returns a bcrypt hash of the OTP.
func HashOTP(otp string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	return string(b), err
}

// VerifyOTP compares an OTP with its bcrypt hash.
func VerifyOTP(hash, otp string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(otp)) == nil
}

// NewRefreshToken returns an opaque refresh token and its SHA-256 hash.
func NewRefreshToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(b)
	return token, HashRefreshToken(token), nil
}

// HashRefreshToken hashes a refresh token for storage.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

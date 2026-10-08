package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

var phoneRe = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)

// NormalizePhone validates and canonicalizes an Indian phone number to E.164.
// It accepts 10-digit local numbers, 0-prefixed numbers and +91 numbers.
func NormalizePhone(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	p = strings.ReplaceAll(p, " ", "")
	p = strings.ReplaceAll(p, "-", "")

	switch {
	case strings.HasPrefix(p, "+"):
		// keep as-is
	case strings.HasPrefix(p, "0"):
		p = "+91" + strings.TrimPrefix(p, "0")
	case len(p) == 10:
		p = "+91" + p
	default:
		if !strings.HasPrefix(p, "+") {
			p = "+" + p
		}
	}

	if !phoneRe.MatchString(p) {
		return "", errors.New("invalid phone number")
	}
	return p, nil
}

// HashPhone returns a stable SHA-256 HMAC of the phone number using a server
// pepper. Plaintext phone numbers are never stored.
func HashPhone(phone, pepper string) string {
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(phone))
	return hex.EncodeToString(mac.Sum(nil))
}

// Last4 returns the last four digits of a normalized phone number for display.
func Last4(phone string) string {
	digits := strings.TrimLeft(phone, "+")
	if len(digits) <= 4 {
		return digits
	}
	return digits[len(digits)-4:]
}

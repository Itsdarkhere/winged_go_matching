package setting

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func normalizeEmail(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return norm.NFC.String(s)
}

func normalizeNumber(s string) string {
	// Adjust to your rule (e.g., E.164). Here: trim + NFC.
	s = strings.TrimSpace(s)
	return norm.NFC.String(s)
}

// StableID returns base64url(HMAC-SHA256(secret, "v1\0"+email+"\0"+number)).
func StableID(secret []byte, email, number string) string {
	e := normalizeEmail(email)
	n := normalizeNumber(number)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("v1"))
	mac.Write([]byte{0})
	mac.Write([]byte(e))
	mac.Write([]byte{0})
	mac.Write([]byte(n))

	sum := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}

package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GenerateTOTPSecret creates a crypto-random base32-encoded 20-byte secret
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// GenerateOTPAuthURL generates an otpauth:// URL suitable for QR code generators and authenticator apps
func GenerateOTPAuthURL(issuer, accountName, secret string) string {
	v := url.Values{}
	v.Set("secret", secret)
	v.Set("issuer", issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	v.Set("period", "30")

	label := url.PathEscape(issuer + ":" + accountName)
	return fmt.Sprintf("otpauth://totp/%s?%s", label, v.Encode())
}

// GenerateCode calculates the 6-digit TOTP code for a given secret and timestamp
func GenerateCode(secret string, t time.Time) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		// try with padding
		key, err = base32.StdEncoding.DecodeString(secret)
		if err != nil {
			return "", fmt.Errorf("invalid base32 secret: %w", err)
		}
	}

	counter := uint64(t.Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(h[offset : offset+4])
	truncatedHash &= 0x7fffffff

	code := truncatedHash % 1000000
	return fmt.Sprintf("%06d", code), nil
}

// VerifyCode verifies a user-provided 6-digit TOTP code with +/- 1 time step tolerance (90s window)
func VerifyCode(secret, code string, t time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}

	// Check window -1, 0, +1
	for _, offset := range []time.Duration{-30 * time.Second, 0, 30 * time.Second} {
		expected, err := GenerateCode(secret, t.Add(offset))
		if err == nil && hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
}

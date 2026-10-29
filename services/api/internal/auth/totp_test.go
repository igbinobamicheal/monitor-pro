package auth_test

import (
	"testing"
	"time"

	"monitopro/services/api/internal/auth"
)

func TestTOTPGenerationAndVerification(t *testing.T) {
	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("Failed to generate TOTP secret: %v", err)
	}
	if len(secret) < 16 {
		t.Fatalf("Secret too short: %s", secret)
	}

	now := time.Now().UTC()
	code, err := auth.GenerateCode(secret, now)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("Expected 6-digit code, got: %s", code)
	}

	// Verify valid code
	if !auth.VerifyCode(secret, code, now) {
		t.Fatalf("Verification failed for valid code: %s", code)
	}

	// Verify code within +/- 30s window
	if !auth.VerifyCode(secret, code, now.Add(25*time.Second)) {
		t.Fatalf("Verification failed for code within drift window")
	}

	// Verify invalid code rejected
	if auth.VerifyCode(secret, "000000", now) && code != "000000" {
		t.Fatalf("Accepted invalid code")
	}

	// Verify malformed code rejected
	if auth.VerifyCode(secret, "abc", now) {
		t.Fatalf("Accepted malformed non-digit code")
	}
}

func TestGenerateOTPAuthURL(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	url := auth.GenerateOTPAuthURL("MonitorPro", "admin@company.com", secret)
	if url == "" {
		t.Fatal("Expected non-empty OTP auth URL")
	}
}

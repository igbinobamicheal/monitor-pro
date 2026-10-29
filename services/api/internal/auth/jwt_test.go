package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "dev_jwt_secret_key_minimum_32_characters_long"
	tm := auth.NewTokenManager(secret, 1*time.Hour)

	userID := uuid.New()
	orgID := uuid.New()
	role := domain.RoleOwnerAdmin
	email := "owner@monitopro.com"

	tokenStr, exp, err := tm.GenerateToken(userID, orgID, role, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("Expected non-empty token string")
	}
	if exp.Before(time.Now().UTC()) {
		t.Fatal("ExpiresAt should be in the future")
	}

	claims, err := tm.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %v, expected %v", claims.UserID, userID)
	}
	if claims.OrgID != orgID {
		t.Errorf("OrgID mismatch: got %v, expected %v", claims.OrgID, orgID)
	}
	if claims.Role != role {
		t.Errorf("Role mismatch: got %v, expected %v", claims.Role, role)
	}
	if claims.Email != email {
		t.Errorf("Email mismatch: got %v, expected %v", claims.Email, email)
	}
}

func TestJWTExpiredAndInvalidSignature(t *testing.T) {
	secret1 := "dev_jwt_secret_key_minimum_32_characters_long_1"
	secret2 := "dev_jwt_secret_key_minimum_32_characters_long_2"

	tm1 := auth.NewTokenManager(secret1, -1*time.Minute) // Already expired
	tm2 := auth.NewTokenManager(secret2, 1*time.Hour)

	userID := uuid.New()
	orgID := uuid.New()

	// 1. Expired token validation should fail
	expiredToken, _, _ := tm1.GenerateToken(userID, orgID, domain.RoleITAdmin, "user@test.com")
	if _, err := tm1.ValidateToken(expiredToken); err == nil {
		t.Fatal("Expected error validating expired token, got nil")
	}

	// 2. Token signed with secret1 validated against secret2 should fail
	validToken, _, _ := auth.NewTokenManager(secret1, 1*time.Hour).GenerateToken(userID, orgID, domain.RoleITAdmin, "user@test.com")
	if _, err := tm2.ValidateToken(validToken); err == nil {
		t.Fatal("Expected error validating token with incorrect signature, got nil")
	}
}

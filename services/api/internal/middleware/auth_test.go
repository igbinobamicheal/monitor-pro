package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
)

func TestAuthenticatorMiddleware(t *testing.T) {
	secret := "test_secret_32_characters_minimum_len!"
	tokenManager := auth.NewTokenManager(secret, 1*time.Hour)
	authenticator := NewAuthenticator(tokenManager)

	userID := uuid.New()
	orgID := uuid.New()
	role := domain.RoleITAdmin
	email := "operator@companya.com"

	token, _, err := tokenManager.GenerateToken(userID, orgID, role, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		claims := UserFromContext(r.Context())
		if claims == nil {
			t.Fatal("Expected claims in context")
		}
		if claims.UserID != userID {
			t.Errorf("Expected user ID %v, got %v", userID, claims.UserID)
		}
		if claims.OrgID != orgID {
			t.Errorf("Expected org ID %v, got %v", orgID, claims.OrgID)
		}
		if claims.Role != role {
			t.Errorf("Expected role %v, got %v", role, claims.Role)
		}
		w.WriteHeader(http.StatusOK)
	})

	protected := authenticator.Authenticate(testHandler)

	req := httptest.NewRequest("GET", "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}
	if !handlerCalled {
		t.Fatal("Handler was not called")
	}
}

func TestRequireRoles(t *testing.T) {
	secret := "test_secret_32_characters_minimum_len!"
	tokenManager := auth.NewTokenManager(secret, 1*time.Hour)
	authenticator := NewAuthenticator(tokenManager)

	userID := uuid.New()
	orgID := uuid.New()

	viewerToken, _, _ := tokenManager.GenerateToken(userID, orgID, domain.RoleReadonlyViewer, "viewer@companya.com")

	protectedHandler := RequireRoles(domain.RoleOwnerAdmin, domain.RoleITAdmin)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	fullChain := authenticator.Authenticate(protectedHandler)

	req := httptest.NewRequest("POST", "/api/v1/remote-support/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+viewerToken)
	rec := httptest.NewRecorder()

	fullChain.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden for ReadonlyViewer, got %d", rec.Code)
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

type AuthHandler struct {
	store  storage.Storage
	tokens *auth.TokenManager
}

func NewAuthHandler(store storage.Storage, tokens *auth.TokenManager) *AuthHandler {
	return &AuthHandler{
		store:  store,
		tokens: tokens,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MFACode  string `json:"mfa_code"`
}

type LoginResponse struct {
	Token     string             `json:"token"`
	ExpiresAt string             `json:"expires_at"`
	User      domain.UserProfile `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email and password are required", "MISSING_CREDENTIALS")
		return
	}

	user, err := h.store.Users().GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid email or password", "INVALID_CREDENTIALS")
		return
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid email or password", "INVALID_CREDENTIALS")
		return
	}

	// Check MFA
	if user.MFAEnabled {
		if req.MFACode == "" {
			writeError(w, http.StatusUnauthorized, "MFA code required", "MFA_REQUIRED")
			return
		}
		if user.MFASecret == nil || !auth.VerifyCode(*user.MFASecret, req.MFACode, time.Now().UTC()) {
			writeError(w, http.StatusUnauthorized, "Invalid MFA code", "INVALID_MFA")
			return
		}
	}

	// Determine organization and role
	var orgID uuid.UUID
	role := domain.RoleReadonlyViewer

	tenant := middleware.TenantFromContext(r.Context())
	if tenant != nil {
		orgID = tenant.ID
		mem, err := h.store.Users().GetMembership(r.Context(), tenant.ID, user.ID)
		if err == nil {
			role = mem.Role
		} else if user.IsPlatformAdmin {
			role = domain.RoleOwnerAdmin
		} else {
			writeError(w, http.StatusForbidden, "User does not have access to this organization", "NO_MEMBERSHIP")
			return
		}
	} else {
		// Public login without tenant context - check if platform admin or fallback default
		if user.IsPlatformAdmin {
			orgID = uuid.Nil
			role = domain.RoleOwnerAdmin
		} else {
			writeError(w, http.StatusBadRequest, "Organization context required", "TENANT_REQUIRED")
			return
		}
	}

	tokenStr, expiresAt, err := h.tokens.GenerateToken(user.ID, orgID, role, user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	profile := domain.UserProfile{
		ID:         user.ID,
		Email:      user.Email,
		FullName:   user.FullName,
		Role:       role,
		MFAEnabled: user.MFAEnabled,
	}

	writeJSON(w, http.StatusOK, LoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      profile,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Session terminated
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) MFASetup(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.UserFromContext(r.Context())
	if userClaims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate secret", "INTERNAL_ERROR")
		return
	}

	// Save secret provisionally (unverified)
	if err := h.store.Users().UpdateMFA(r.Context(), userClaims.UserID, &secret, false); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save MFA configuration", "INTERNAL_ERROR")
		return
	}

	qrURL := auth.GenerateOTPAuthURL("MonitorPro", userClaims.Email, secret)
	writeJSON(w, http.StatusOK, map[string]string{
		"secret":      secret,
		"qr_code_url": qrURL,
	})
}

type MFAVerifyRequest struct {
	Code string `json:"code"`
}

func (h *AuthHandler) MFAVerify(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.UserFromContext(r.Context())
	if userClaims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	user, err := h.store.Users().GetByID(r.Context(), userClaims.UserID)
	if err != nil || user.MFASecret == nil {
		writeError(w, http.StatusBadRequest, "MFA setup has not been initiated", "MFA_NOT_INITIATED")
		return
	}

	if !auth.VerifyCode(*user.MFASecret, req.Code, time.Now().UTC()) {
		writeError(w, http.StatusBadRequest, "Invalid verification code", "INVALID_CODE")
		return
	}

	if err := h.store.Users().UpdateMFA(r.Context(), user.ID, user.MFASecret, true); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to activate MFA", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "MFA activated",
	})
}

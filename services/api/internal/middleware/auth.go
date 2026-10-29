package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
)

// Authenticator validates Bearer JWT tokens and checks tenant alignment
type Authenticator struct {
	tokens *auth.TokenManager
}

func NewAuthenticator(tokens *auth.TokenManager) *Authenticator {
	return &Authenticator{tokens: tokens}
}

// Authenticate verifies the JWT token in Authorization header
func (a *Authenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(domain.ErrorResponse{
				Error: "Missing Authorization header",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(domain.ErrorResponse{
				Error: "Invalid Authorization header format; expected Bearer <token>",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		claims, err := a.tokens.ValidateToken(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(domain.ErrorResponse{
				Error: "Invalid or expired token",
				Code:  "INVALID_TOKEN",
			})
			return
		}

		// Enforce tenant isolation if tenant is in context
		tenant := TenantFromContext(r.Context())
		if tenant != nil && tenant.ID != claims.OrgID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(domain.ErrorResponse{
				Error: "User does not belong to active tenant",
				Code:  "TENANT_MISMATCH",
			})
			return
		}

		ctx := WithUser(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoles enforces that the authenticated user possesses one of the allowed roles
func RequireRoles(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]bool)
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(domain.ErrorResponse{
					Error: "Authentication required",
					Code:  "UNAUTHORIZED",
				})
				return
			}

			if !allowed[user.Role] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(domain.ErrorResponse{
					Error: "Forbidden: insufficient permissions for this action",
					Code:  "FORBIDDEN",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

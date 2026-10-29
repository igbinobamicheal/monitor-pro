package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/storage"
)

// ExtractSubdomain extracts tenant identifier from Host header
func ExtractSubdomain(host string) string {
	// Strip port if present
	h, _, err := net.SplitHostPort(host)
	if err == nil {
		host = h
	}

	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return ""
	}

	sub := strings.ToLower(parts[0])
	// Filter out non-tenant subdomains
	switch sub {
	case "localhost", "127", "api", "www", "app", "monitopro":
		return ""
	default:
		return sub
	}
}

// TenantResolver handles tenant extraction and resolution from DB
type TenantResolver struct {
	tenants storage.TenantRepository
}

func NewTenantResolver(tenants storage.TenantRepository) *TenantResolver {
	return &TenantResolver{tenants: tenants}
}

// ResolveMiddleware inspects request for tenant clues and attaches Organization to context
func (tr *TenantResolver) ResolveMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var org *domain.Organization

		// 1. Check X-Tenant-ID header
		headerVal := strings.TrimSpace(r.Header.Get(TenantIDHeader))
		if headerVal != "" {
			if id, err := uuid.Parse(headerVal); err == nil {
				org, _ = tr.tenants.GetByID(ctx, id)
			} else {
				org, _ = tr.tenants.GetBySubdomain(ctx, headerVal)
			}
		}

		// 2. If not found in header, check Host subdomain
		if org == nil {
			subdomain := ExtractSubdomain(r.Host)
			if subdomain != "" {
				org, _ = tr.tenants.GetBySubdomain(ctx, subdomain)
			}
		}

		if org != nil {
			ctx = WithTenant(ctx, org)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireTenant ensures that a valid tenant has been resolved, rejecting if unknown
func (tr *TenantResolver) RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org := TenantFromContext(r.Context())
		if org == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(domain.ErrorResponse{
				Error: "Tenant not found or invalid",
				Code:  "TENANT_NOT_FOUND",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

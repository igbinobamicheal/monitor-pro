package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"company.monitopro.com", "company"},
		{"acme.localhost:8080", "acme"},
		{"tenant1.monitopro.localhost", "tenant1"},
		{"localhost:8080", ""},
		{"127.0.0.1:8080", ""},
		{"api.monitopro.com", ""},
		{"monitopro.com", ""},
		{"www.monitopro.com", ""},
	}

	for _, tt := range tests {
		got := middleware.ExtractSubdomain(tt.host)
		if got != tt.expected {
			t.Errorf("ExtractSubdomain(%q) = %q; expected %q", tt.host, got, tt.expected)
		}
	}
}

func TestTenantResolverMiddleware(t *testing.T) {
	store := storage.NewMemoryStorage()
	orgID := uuid.New()
	org := &domain.Organization{
		ID:        orgID,
		Name:      "Acme Corp",
		Subdomain: "acme",
		AIEnabled: true,
	}
	_ = store.Tenants().Create(context.Background(), org)

	resolver := middleware.NewTenantResolver(store.Tenants())

	// 1. Resolve via Host header
	{
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.Host = "acme.monitopro.com"
		rec := httptest.NewRecorder()

		var resolvedOrg *domain.Organization
		handler := resolver.ResolveMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resolvedOrg = middleware.TenantFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		if resolvedOrg == nil || resolvedOrg.ID != orgID {
			t.Fatalf("Failed to resolve tenant from Host header: %v", resolvedOrg)
		}
	}

	// 2. Resolve via X-Tenant-ID header (UUID)
	{
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.Header.Set(middleware.TenantIDHeader, orgID.String())
		rec := httptest.NewRecorder()

		var resolvedOrg *domain.Organization
		handler := resolver.ResolveMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resolvedOrg = middleware.TenantFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		if resolvedOrg == nil || resolvedOrg.ID != orgID {
			t.Fatalf("Failed to resolve tenant from X-Tenant-ID UUID header")
		}
	}

	// 3. Resolve via X-Tenant-ID header (Subdomain)
	{
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		req.Header.Set(middleware.TenantIDHeader, "acme")
		rec := httptest.NewRecorder()

		var resolvedOrg *domain.Organization
		handler := resolver.ResolveMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resolvedOrg = middleware.TenantFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		if resolvedOrg == nil || resolvedOrg.ID != orgID {
			t.Fatalf("Failed to resolve tenant from X-Tenant-ID subdomain header")
		}
	}

	// 4. RequireTenant rejects when no tenant resolved
	{
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		rec := httptest.NewRecorder()

		chain := resolver.ResolveMiddleware(resolver.RequireTenant(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))

		chain.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Expected status 404 for unresolvable tenant, got %d", rec.Code)
		}
	}
}

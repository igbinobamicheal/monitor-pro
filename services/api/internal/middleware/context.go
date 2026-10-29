package middleware

import (
	"context"

	"github.com/google/uuid"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
)

type contextKey string

const (
	TenantContextKey  contextKey = "monitopro_tenant"
	UserContextKey    contextKey = "monitopro_user"
	RequestIDKey      contextKey = "monitopro_request_id"
	TenantIDHeader               = "X-Tenant-ID"
	RequestIDHeader              = "X-Request-ID"
	AgentSignatureHdr            = "X-Agent-Signature"
	DeviceIDHeader               = "X-Device-ID"
)

// WithTenant stores organization in context
func WithTenant(ctx context.Context, org *domain.Organization) context.Context {
	return context.WithValue(ctx, TenantContextKey, org)
}

// TenantFromContext retrieves organization from context
func TenantFromContext(ctx context.Context) *domain.Organization {
	val := ctx.Value(TenantContextKey)
	if val == nil {
		return nil
	}
	if org, ok := val.(*domain.Organization); ok {
		return org
	}
	return nil
}

// TenantIDFromContext returns the organization ID or uuid.Nil
func TenantIDFromContext(ctx context.Context) uuid.UUID {
	org := TenantFromContext(ctx)
	if org != nil {
		return org.ID
	}
	return uuid.Nil
}

// WithUser stores JWT claims in context
func WithUser(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}

// UserFromContext retrieves JWT claims from context
func UserFromContext(ctx context.Context) *auth.Claims {
	val := ctx.Value(UserContextKey)
	if val == nil {
		return nil
	}
	if claims, ok := val.(*auth.Claims); ok {
		return claims
	}
	return nil
}

// WithRequestID stores request ID in context
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, reqID)
}

// RequestIDFromContext retrieves request ID from context
func RequestIDFromContext(ctx context.Context) string {
	val := ctx.Value(RequestIDKey)
	if val == nil {
		return ""
	}
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

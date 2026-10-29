package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
)

// Claims represents our standard JWT payload
type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	OrgID  uuid.UUID   `json:"org_id"`
	Role   domain.Role `json:"role"`
	Email  string      `json:"email"`
	jwt.RegisteredClaims
}

// TokenManager handles generation and validation of JWT tokens
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager initializes TokenManager
func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// GenerateToken generates a signed JWT string for the given user, organization, and role
func (m *TokenManager) GenerateToken(userID, orgID uuid.UUID, role domain.Role, email string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := &Claims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    "monitopro-api",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, expiresAt, nil
}

// ValidateToken parses and validates a JWT token string
func (m *TokenManager) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

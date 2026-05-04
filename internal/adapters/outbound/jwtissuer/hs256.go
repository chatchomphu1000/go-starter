// Package jwtissuer provides an HS256 JWT implementation of the TokenIssuer port.
package jwtissuer

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
)

// HS256Issuer implements outbound.TokenIssuer using HMAC-SHA256.
type HS256Issuer struct {
	secret     []byte
	ttl        time.Duration
	refreshTTL time.Duration
	issuer     string
	clock      outbound.Clock
}

// NewHS256Issuer creates a new HS256 TokenIssuer.
// Panics if secret is less than 32 bytes.
// Refresh token TTL is set to 7× the access token TTL.
func NewHS256Issuer(secret []byte, ttl time.Duration, issuer string, clock outbound.Clock) *HS256Issuer {
	if len(secret) < 32 {
		panic("jwtissuer: secret must be at least 32 bytes")
	}
	return &HS256Issuer{
		secret:     secret,
		ttl:        ttl,
		refreshTTL: ttl * 8,
		issuer:     issuer,
		clock:      clock,
	}
}

// Issue creates a new signed JWT for the given user ID and role.
func (i *HS256Issuer) Issue(_ context.Context, userID string, role domain.Role) (string, time.Time, error) {
	now := i.clock.Now()
	expiresAt := now.Add(i.ttl)

	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role.String(),
		"iat":  now.Unix(),
		"exp":  expiresAt.Unix(),
		"jti":  uuid.New().String(),
		"iss":  i.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("hs256Issuer.Issue: %w", err)
	}

	return signed, expiresAt, nil
}

// IssueRefresh creates a new signed refresh JWT for the given user ID and role.
// Refresh tokens have a longer TTL and carry a "type":"refresh" claim.
func (i *HS256Issuer) IssueRefresh(_ context.Context, userID string, role domain.Role) (string, time.Time, error) {
	now := i.clock.Now()
	expiresAt := now.Add(i.refreshTTL)

	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role.String(),
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  expiresAt.Unix(),
		"jti":  uuid.New().String(),
		"iss":  i.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("hs256Issuer.IssueRefresh: %w", err)
	}

	return signed, expiresAt, nil
}

// VerifyRefresh validates a refresh JWT and returns the user ID and role.
// Returns an error if the token is not a refresh token.
func (i *HS256Issuer) VerifyRefresh(_ context.Context, tokenString string) (string, domain.Role, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secret, nil
	}, jwt.WithIssuer(i.issuer), jwt.WithExpirationRequired())

	if err != nil {
		return "", "", apperrors.Unauthorized("invalid or expired refresh token", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", apperrors.Unauthorized("invalid refresh token claims", nil)
	}

	if tokenType, _ := claims["type"].(string); tokenType != "refresh" {
		return "", "", apperrors.Unauthorized("token is not a refresh token", nil)
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", "", apperrors.Unauthorized("missing subject in refresh token", nil)
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return "", "", apperrors.Unauthorized("missing role in refresh token", nil)
	}

	role, err := domain.ParseRole(roleStr)
	if err != nil {
		return "", "", apperrors.Unauthorized("invalid role in refresh token", err)
	}

	return sub, role, nil
}

// Verify validates a JWT and returns the user ID and role.
func (i *HS256Issuer) Verify(_ context.Context, tokenString string) (string, domain.Role, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secret, nil
	}, jwt.WithIssuer(i.issuer), jwt.WithExpirationRequired())

	if err != nil {
		return "", "", apperrors.Unauthorized("invalid or expired token", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", apperrors.Unauthorized("invalid token claims", nil)
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", "", apperrors.Unauthorized("missing subject in token", nil)
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return "", "", apperrors.Unauthorized("missing role in token", nil)
	}

	role, err := domain.ParseRole(roleStr)
	if err != nil {
		return "", "", apperrors.Unauthorized("invalid role in token", err)
	}

	return sub, role, nil
}

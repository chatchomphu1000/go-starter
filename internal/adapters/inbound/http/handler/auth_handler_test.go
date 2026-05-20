package handler_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/handler"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
)

func validAuthToken() *inbound.AuthToken {
	return &inbound.AuthToken{
		AccessToken:      "access-tok",
		TokenType:        "Bearer",
		ExpiresAt:        fixedNow.Add(24 * time.Hour),
		RefreshToken:     "refresh-tok",
		RefreshExpiresAt: fixedNow.Add(7 * 24 * time.Hour),
	}
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MockAuthService)
		wantStatus int
		wantInBody string
	}{
		{
			name: "success_201",
			body: `{"name":"Alice","email":"alice@example.com","password":"Secret@1234"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Register(mock.Anything, mock.Anything).Return(validDomainUser(t), nil)
			},
			wantStatus: http.StatusCreated,
			wantInBody: `"id":"user-1"`,
		},
		{
			name:       "missing_required_fields_400",
			body:       `{"email":"alice@example.com"}`,
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_json_400",
			body:       `not-json`,
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "email_conflict_409",
			body: `{"name":"Alice","email":"alice@example.com","password":"Secret@1234"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Register(mock.Anything, mock.Anything).
					Return(nil, domain.ErrEmailAlreadyExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "response_no_hashed_password",
			body: `{"name":"Alice","email":"alice@example.com","password":"Secret@1234"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Register(mock.Anything, mock.Anything).Return(validDomainUser(t), nil)
			},
			wantStatus: http.StatusCreated,
			wantInBody: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockAuthService)
			tc.setup(svc)

			h := handler.NewAuthHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodPost, "/api/v1/auth/register", tc.body)
			call(e, c, h.Register)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantInBody != "" {
				assert.Contains(t, rec.Body.String(), tc.wantInBody)
			}
			assert.NotContains(t, rec.Body.String(), "hashed_password")
			svc.AssertExpectations(t)
		})
	}
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success_200",
			body: `{"email":"alice@example.com","password":"Secret@1234"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Login(mock.Anything, mock.Anything).Return(validAuthToken(), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing_fields_400",
			body:       `{"email":"alice@example.com"}`,
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_credentials_401",
			body: `{"email":"alice@example.com","password":"wrongpassword"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Login(mock.Anything, mock.Anything).
					Return(nil, domain.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "inactive_user_403",
			body: `{"email":"alice@example.com","password":"Secret@1234"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Login(mock.Anything, mock.Anything).
					Return(nil, domain.ErrUserInactive)
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockAuthService)
			tc.setup(svc)

			h := handler.NewAuthHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodPost, "/api/v1/auth/login", tc.body)
			call(e, c, h.Login)

			assert.Equal(t, tc.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestAuthHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success_200",
			body: `{"refresh_token":"valid-refresh"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().RefreshToken(mock.Anything, mock.Anything).Return(validAuthToken(), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing_refresh_token_400",
			body:       `{}`,
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid_token_401",
			body: `{"refresh_token":"expired"}`,
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().RefreshToken(mock.Anything, mock.Anything).
					Return(nil, domain.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockAuthService)
			tc.setup(svc)

			h := handler.NewAuthHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodPost, "/api/v1/auth/refresh", tc.body)
			call(e, c, h.RefreshToken)

			require.Equal(t, tc.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

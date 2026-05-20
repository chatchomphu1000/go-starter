package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

var fixedNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

func nopLogger() logger.Logger {
	l, _ := logger.NewLogger(logger.LoggerConfig{Level: "error", Format: "json"})
	return l
}

func cancelledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

type authMocks struct {
	repo     *mocks.MockUserRepository
	notifier *mocks.MockNotifier
	hasher   *mocks.MockPasswordHasher
	tokens   *mocks.MockTokenIssuer
	clock    *mocks.MockClock
	ids      *mocks.MockIDGenerator
}

func newTestAuthService(t *testing.T) (*authService, authMocks) {
	t.Helper()
	m := authMocks{
		repo:     new(mocks.MockUserRepository),
		notifier: new(mocks.MockNotifier),
		hasher:   new(mocks.MockPasswordHasher),
		tokens:   new(mocks.MockTokenIssuer),
		clock:    new(mocks.MockClock),
		ids:      new(mocks.MockIDGenerator),
	}
	svc := NewAuthService(m.repo, m.notifier, m.hasher, m.tokens, m.clock, m.ids, nopLogger()).(*authService)
	return svc, m
}

func validUser(t *testing.T) *domain.User {
	t.Helper()
	email, _ := domain.NewEmail("alice@example.com")
	u, _ := domain.NewUser("user-1", "Alice", email, "hashed", domain.RoleUser, fixedNow)
	return u
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestAuthService_Register(t *testing.T) {
	validInput := inbound.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "Secret@1234",
	}

	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.RegisterInput
		setup   func(m authMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				m.repo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(false, nil)
				m.hasher.EXPECT().Hash("Secret@1234").Return("hashed", nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.ids.EXPECT().New().Return("user-1")
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
				// notifier is called asynchronously — not asserted here
				m.notifier.EXPECT().SendWelcomeEmail(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: nil,
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validInput,
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "weak_password_too_short",
			ctx:   context.Background(),
			input: inbound.RegisterInput{Name: "Alice", Email: "alice@example.com", Password: "Short1!"},
			setup: func(_ authMocks) {},
			// no DB calls expected
			wantErr: domain.ErrWeakPassword,
		},
		{
			name:  "weak_password_no_upper",
			ctx:   context.Background(),
			input: inbound.RegisterInput{Name: "Alice", Email: "alice@example.com", Password: "secret@1234"},
			setup: func(_ authMocks) {},
			wantErr: domain.ErrWeakPassword,
		},
		{
			name:  "weak_password_no_symbol",
			ctx:   context.Background(),
			input: inbound.RegisterInput{Name: "Alice", Email: "alice@example.com", Password: "Secret12345"},
			setup: func(_ authMocks) {},
			wantErr: domain.ErrWeakPassword,
		},
		{
			name:  "invalid_email_format",
			ctx:   context.Background(),
			input: inbound.RegisterInput{Name: "Alice", Email: "not-an-email", Password: "Secret@1234"},
			setup: func(_ authMocks) {},
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name:  "email_already_exists",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				m.repo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(true, nil)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:  "hash_error",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				m.repo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(false, nil)
				m.hasher.EXPECT().Hash("Secret@1234").Return("", errors.New("bcrypt error"))
			},
			wantErr: errors.New("bcrypt error"),
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				m.repo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(false, nil)
				m.hasher.EXPECT().Hash("Secret@1234").Return("hashed", nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.ids.EXPECT().New().Return("user-1")
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			got, err := svc.Register(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				// For sentinel domain errors, use errors.Is
				if errors.Is(tc.wantErr, domain.ErrWeakPassword) ||
					errors.Is(tc.wantErr, domain.ErrInvalidEmail) ||
					errors.Is(tc.wantErr, domain.ErrEmailAlreadyExists) {
					assert.True(t, errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "user-1", got.ID)
			}

			m.repo.AssertExpectations(t)
			m.hasher.AssertExpectations(t)
			m.clock.AssertExpectations(t)
			m.ids.AssertExpectations(t)
		})
	}
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestAuthService_Login(t *testing.T) {
	validInput := inbound.LoginInput{
		Email:    "alice@example.com",
		Password: "Secret@1234",
	}
	expiresAt := fixedNow.Add(24 * time.Hour)
	refreshExp := fixedNow.Add(7 * 24 * time.Hour)

	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.LoginInput
		setup   func(m authMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				u := validUser(t)
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
				m.hasher.EXPECT().Verify(u.HashedPassword, "Secret@1234").Return(nil)
				m.tokens.EXPECT().Issue(mock.Anything, u.ID, u.Role).Return("access-token", expiresAt, nil)
				m.tokens.EXPECT().IssueRefresh(mock.Anything, u.ID, u.Role).Return("refresh-token", refreshExp, nil)
			},
			wantErr: nil,
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validInput,
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "invalid_email_format",
			ctx:   context.Background(),
			input: inbound.LoginInput{Email: "bad-email", Password: "Secret@1234"},
			setup: func(_ authMocks) {},
			// must return ErrInvalidCredentials, NOT ErrInvalidEmail (prevents user enumeration)
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "user_not_found",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "wrong_password",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				u := validUser(t)
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
				m.hasher.EXPECT().Verify(u.HashedPassword, "Secret@1234").Return(domain.ErrInvalidCredentials)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "user_inactive",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				u := validUser(t)
				u.Active = false
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
				m.hasher.EXPECT().Verify(u.HashedPassword, "Secret@1234").Return(nil)
			},
			wantErr: domain.ErrUserInactive,
		},
		{
			name:  "token_issue_error",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				u := validUser(t)
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
				m.hasher.EXPECT().Verify(u.HashedPassword, "Secret@1234").Return(nil)
				m.tokens.EXPECT().Issue(mock.Anything, u.ID, u.Role).Return("", time.Time{}, errors.New("jwt error"))
			},
			wantErr: errors.New("jwt error"),
		},
		{
			name:  "refresh_issue_error",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m authMocks) {
				u := validUser(t)
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
				m.hasher.EXPECT().Verify(u.HashedPassword, "Secret@1234").Return(nil)
				m.tokens.EXPECT().Issue(mock.Anything, u.ID, u.Role).Return("access-token", expiresAt, nil)
				m.tokens.EXPECT().IssueRefresh(mock.Anything, u.ID, u.Role).Return("", time.Time{}, errors.New("refresh error"))
			},
			wantErr: errors.New("refresh error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			got, err := svc.Login(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrInvalidCredentials) ||
					errors.Is(tc.wantErr, domain.ErrUserInactive) {
					assert.True(t, errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "access-token", got.AccessToken)
				assert.Equal(t, "Bearer", got.TokenType)
				assert.Equal(t, "refresh-token", got.RefreshToken)
			}

			m.repo.AssertExpectations(t)
			m.hasher.AssertExpectations(t)
			m.tokens.AssertExpectations(t)
		})
	}
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestAuthService_RefreshToken(t *testing.T) {
	expiresAt := fixedNow.Add(24 * time.Hour)
	refreshExp := fixedNow.Add(7 * 24 * time.Hour)

	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.RefreshTokenInput
		setup   func(m authMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: inbound.RefreshTokenInput{RefreshToken: "valid-refresh"},
			setup: func(m authMocks) {
				u := validUser(t)
				m.tokens.EXPECT().VerifyRefresh(mock.Anything, "valid-refresh").Return(u.ID, u.Role, nil)
				m.repo.EXPECT().FindByID(mock.Anything, u.ID).Return(u, nil)
				m.tokens.EXPECT().Issue(mock.Anything, u.ID, u.Role).Return("new-access", expiresAt, nil)
				m.tokens.EXPECT().IssueRefresh(mock.Anything, u.ID, u.Role).Return("new-refresh", refreshExp, nil)
			},
			wantErr: nil,
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   inbound.RefreshTokenInput{RefreshToken: "valid-refresh"},
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "invalid_refresh_token",
			ctx:   context.Background(),
			input: inbound.RefreshTokenInput{RefreshToken: "bad-token"},
			setup: func(m authMocks) {
				m.tokens.EXPECT().VerifyRefresh(mock.Anything, "bad-token").Return("", domain.Role(""), errors.New("invalid token"))
			},
			wantErr: errors.New("invalid token"),
		},
		{
			name:  "user_not_found",
			ctx:   context.Background(),
			input: inbound.RefreshTokenInput{RefreshToken: "valid-refresh"},
			setup: func(m authMocks) {
				m.tokens.EXPECT().VerifyRefresh(mock.Anything, "valid-refresh").Return("user-1", domain.RoleUser, nil)
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "user_inactive",
			ctx:   context.Background(),
			input: inbound.RefreshTokenInput{RefreshToken: "valid-refresh"},
			setup: func(m authMocks) {
				u := validUser(t)
				u.Active = false
				m.tokens.EXPECT().VerifyRefresh(mock.Anything, "valid-refresh").Return(u.ID, u.Role, nil)
				m.repo.EXPECT().FindByID(mock.Anything, u.ID).Return(u, nil)
			},
			wantErr: domain.ErrUserInactive,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			got, err := svc.RefreshToken(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrInvalidCredentials) || errors.Is(tc.wantErr, domain.ErrUserInactive) {
					assert.True(t, errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, "new-access", got.AccessToken)
				assert.Equal(t, "new-refresh", got.RefreshToken)
			}

			m.repo.AssertExpectations(t)
			m.tokens.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestAuthService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m authMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "user-1",
			setup: func(m authMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUser(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "user-1",
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m authMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			got, err := svc.GetByID(tc.ctx, tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestAuthService_List(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		filter     inbound.ListFilter
		setup      func(m authMocks)
		wantPage   int
		wantLimit  int
		wantErr    error
	}{
		{
			name:   "success_default_pagination",
			ctx:    context.Background(),
			filter: inbound.ListFilter{Page: 0, Limit: 0},
			setup: func(m authMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.ListFilter) bool {
					return f.Page == 1 && f.Limit == 10
				})).Return([]*domain.User{validUser(t)}, int64(1), nil)
			},
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:   "success_limit_clamped_to_100",
			ctx:    context.Background(),
			filter: inbound.ListFilter{Page: 1, Limit: 200},
			setup: func(m authMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.ListFilter) bool {
					return f.Limit == 100
				})).Return([]*domain.User{}, int64(0), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.ListFilter{},
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.ListFilter{Page: 1, Limit: 10},
			setup: func(m authMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			got, total, err := svc.List(tc.ctx, tc.filter)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				_ = total
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func TestAuthService_Update(t *testing.T) {
	newName := "Alice Updated"

	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		input   inbound.UpdateInput
		setup   func(m authMocks)
		wantErr error
	}{
		{
			name:  "success_rename",
			ctx:   context.Background(),
			id:    "user-1",
			input: inbound.UpdateInput{Name: &newName},
			setup: func(m authMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUser(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:  "success_no_changes",
			ctx:   context.Background(),
			id:    "user-1",
			input: inbound.UpdateInput{Name: nil},
			setup: func(m authMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUser(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "user-1",
			input:   inbound.UpdateInput{Name: &newName},
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "user_not_found",
			ctx:   context.Background(),
			id:    "missing",
			input: inbound.UpdateInput{Name: &newName},
			setup: func(m authMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			got, err := svc.Update(tc.ctx, tc.id, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestAuthService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m authMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "user-1",
			setup: func(m authMocks) {
				m.repo.EXPECT().Delete(mock.Anything, "user-1").Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "user-1",
			setup:   func(_ authMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "user_not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m authMocks) {
				m.repo.EXPECT().Delete(mock.Anything, "missing").Return(domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestAuthService(t)
			tc.setup(m)

			err := svc.Delete(tc.ctx, tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

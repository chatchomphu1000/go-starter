package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
)

type userMocks struct {
	repo     *mocks.MockUserRepository
	notifier *mocks.MockNotifier
	hasher   *mocks.MockPasswordHasher
	tokens   *mocks.MockTokenIssuer
	clock    *mocks.MockClock
	ids      *mocks.MockIDGenerator
}

func newTestUserService(t *testing.T) (*userService, userMocks) {
	t.Helper()
	m := userMocks{
		repo:     new(mocks.MockUserRepository),
		notifier: new(mocks.MockNotifier),
		hasher:   new(mocks.MockPasswordHasher),
		tokens:   new(mocks.MockTokenIssuer),
		clock:    new(mocks.MockClock),
		ids:      new(mocks.MockIDGenerator),
	}
	svc := NewUserService(m.repo, m.notifier, m.hasher, m.tokens, m.clock, m.ids, nopLogger()).(*userService)
	return svc, m
}

func validUserSvc(t *testing.T) *domain.User {
	t.Helper()
	email, _ := domain.NewEmail("alice@example.com")
	u, _ := domain.NewUser("user-1", "Alice", email, "hashed", domain.RoleUser, fixedNow)
	return u
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestUserService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m userMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "user-1",
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUserSvc(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "user-1",
			setup:   func(_ userMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestUserService(t)
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

func TestUserService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.ListFilter
		setup   func(m userMocks)
		wantErr error
	}{
		{
			name:   "success_default_pagination",
			ctx:    context.Background(),
			filter: inbound.ListFilter{Page: 0, Limit: 0},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.ListFilter) bool {
					return f.Page == 1 && f.Limit == 10
				})).Return([]*domain.User{validUserSvc(t)}, int64(1), nil)
			},
		},
		{
			name:   "success_limit_clamped_to_100",
			ctx:    context.Background(),
			filter: inbound.ListFilter{Page: 1, Limit: 200},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.ListFilter) bool {
					return f.Limit == 100
				})).Return([]*domain.User{}, int64(0), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.ListFilter{},
			setup:   func(_ userMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.ListFilter{Page: 1, Limit: 10},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestUserService(t)
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

func TestUserService_Update(t *testing.T) {
	newName := "Alice Updated"
	emptyName := ""

	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		input   inbound.UpdateInput
		setup   func(m userMocks)
		wantErr error
	}{
		{
			name:  "success_rename",
			ctx:   context.Background(),
			id:    "user-1",
			input: inbound.UpdateInput{Name: &newName},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUserSvc(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:  "success_no_changes",
			ctx:   context.Background(),
			id:    "user-1",
			input: inbound.UpdateInput{Name: nil},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUserSvc(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "user-1",
			input:   inbound.UpdateInput{Name: &newName},
			setup:   func(_ userMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "user_not_found",
			ctx:   context.Background(),
			id:    "missing",
			input: inbound.UpdateInput{Name: &newName},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:  "empty_name",
			ctx:   context.Background(),
			id:    "user-1",
			input: inbound.UpdateInput{Name: &emptyName},
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "user-1").Return(validUserSvc(t), nil)
			},
			wantErr: errors.New("name cannot be empty"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestUserService(t)
			tc.setup(m)

			got, err := svc.Update(tc.ctx, tc.id, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrUserNotFound) {
					assert.True(t, errors.Is(err, domain.ErrUserNotFound))
				}
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

func TestUserService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m userMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "user-1",
			setup: func(m userMocks) {
				m.repo.EXPECT().Delete(mock.Anything, "user-1").Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "user-1",
			setup:   func(_ userMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "user_not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m userMocks) {
				m.repo.EXPECT().Delete(mock.Anything, "missing").Return(domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestUserService(t)
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

// ─── Login ───────────────────────────────────────────────────────────────────

func TestUserService_Login(t *testing.T) {
	validInput := inbound.LoginInput{
		Email:    "alice@example.com",
		Password: "Secret@1234",
	}

	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.LoginInput
		setup   func(m userMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(validUserSvc(t), nil)
				m.hasher.EXPECT().Verify(mock.Anything, "Secret@1234").Return(nil)
				m.tokens.EXPECT().Issue(mock.Anything, "user-1", mock.Anything).Return("tok", fixedNow, nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validInput,
			setup:   func(_ userMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "invalid_email_format",
			ctx:   context.Background(),
			input: inbound.LoginInput{Email: "not-an-email", Password: "pass"},
			setup: func(_ userMocks) {},
			// anti-enumeration: returns ErrInvalidCredentials, not ErrInvalidEmail
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "user_not_found",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "wrong_password",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(validUserSvc(t), nil)
				m.hasher.EXPECT().Verify(mock.Anything, "Secret@1234").Return(domain.ErrInvalidCredentials)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "user_inactive",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				u := validUserSvc(t)
				u.Active = false
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
				m.hasher.EXPECT().Verify(mock.Anything, "Secret@1234").Return(nil)
			},
			wantErr: domain.ErrUserInactive,
		},
		{
			name:  "token_issue_error",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				m.repo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(validUserSvc(t), nil)
				m.hasher.EXPECT().Verify(mock.Anything, "Secret@1234").Return(nil)
				m.tokens.EXPECT().Issue(mock.Anything, "user-1", mock.Anything).Return("", fixedNow, errors.New("jwt error"))
			},
			wantErr: errors.New("jwt error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestUserService(t)
			tc.setup(m)

			got, err := svc.Login(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrInvalidCredentials) || errors.Is(tc.wantErr, domain.ErrUserInactive) || errors.Is(tc.wantErr, context.Canceled) {
					assert.True(t, errors.Is(err, tc.wantErr))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "Bearer", got.TokenType)
			}
			m.repo.AssertExpectations(t)
			m.hasher.AssertExpectations(t)
			m.tokens.AssertExpectations(t)
		})
	}
}

// ─── Register (userService also has this — mirrors auth) ─────────────────────

func TestUserService_Register(t *testing.T) {
	validInput := inbound.RegisterInput{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "Secret@1234",
	}

	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.RegisterInput
		setup   func(m userMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				m.repo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(false, nil)
				m.hasher.EXPECT().Hash("Secret@1234").Return("hashed", nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.ids.EXPECT().New().Return("user-2")
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
				// notifier is called asynchronously — not asserted here
				m.notifier.EXPECT().SendWelcomeEmail(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validInput,
			setup:   func(_ userMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "email_already_exists",
			ctx:   context.Background(),
			input: validInput,
			setup: func(m userMocks) {
				m.repo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(true, nil)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestUserService(t)
			tc.setup(m)

			got, err := svc.Register(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrEmailAlreadyExists) {
					assert.True(t, errors.Is(err, tc.wantErr))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}

			m.repo.AssertExpectations(t)
			m.hasher.AssertExpectations(t)
		})
	}
}

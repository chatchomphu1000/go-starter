// Package services implements the application's business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"unicode"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// authService implements inbound.AuthService.
type authService struct {
	repo     outbound.UserRepository
	notifier outbound.Notifier
	hasher   outbound.PasswordHasher
	tokens   outbound.TokenIssuer
	clock    outbound.Clock
	ids      outbound.IDGenerator
	log      logger.Logger
}

// NewAuthService creates a new AuthService with all required dependencies.
func NewAuthService(
	repo outbound.UserRepository,
	notifier outbound.Notifier,
	hasher outbound.PasswordHasher,
	tokens outbound.TokenIssuer,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.AuthService {
	return &authService{
		repo:     repo,
		notifier: notifier,
		hasher:   hasher,
		tokens:   tokens,
		clock:    clock,
		ids:      ids,
		log:      log,
	}
}

// Register creates a new user account.
func (s *authService) Register(ctx context.Context, in inbound.RegisterInput) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}

	if err := validatePassword(in.Password); err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}

	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}

	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("authService.Register: %w", domain.ErrEmailAlreadyExists)
	}

	hashed, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}

	now := s.clock.Now()
	id := s.ids.New()

	user, err := domain.NewUser(id, in.Name, email, hashed, domain.RoleUser, now)
	if err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}

	if err := s.repo.Insert(ctx, user); err != nil {
		return nil, fmt.Errorf("authService.Register: %w", err)
	}

	s.log.Info("user registered", zap.String("user_id", user.ID))

	// Send welcome email in background — do not block the response.
	go func() {
		bgCtx := context.WithoutCancel(ctx)
		if notifyErr := s.notifier.SendWelcomeEmail(bgCtx, user.Email, user.Name); notifyErr != nil {
			s.log.Error("failed to send welcome email", zap.String("user_id", user.ID), zap.Error(notifyErr))
		}
	}()

	return user, nil
}

// Login authenticates a user and returns an auth token.
func (s *authService) Login(ctx context.Context, in inbound.LoginInput) (*inbound.AuthToken, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}

	email, err := domain.NewEmail(in.Email)
	if err != nil {
		// Return generic invalid credentials to prevent user enumeration.
		return nil, fmt.Errorf("authService.Login: %w", domain.ErrInvalidCredentials)
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			// Unexpected DB error — log internally but return generic error to client.
			s.log.Error("unexpected error finding user for login", zap.Error(err))
		}
		// Return generic invalid credentials to prevent user enumeration.
		return nil, fmt.Errorf("authService.Login: %w", domain.ErrInvalidCredentials)
	}

	if err := s.hasher.Verify(user.HashedPassword, in.Password); err != nil {
		s.log.Debug("login failed: password mismatch", zap.String("user_id", user.ID))
		return nil, fmt.Errorf("authService.Login: %w", domain.ErrInvalidCredentials)
	}

	if !user.Active {
		return nil, fmt.Errorf("authService.Login: %w", domain.ErrUserInactive)
	}

	token, expiresAt, err := s.tokens.Issue(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.tokens.IssueRefresh(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("authService.Login: %w", err)
	}

	s.log.Info("user logged in", zap.String("user_id", user.ID))

	return &inbound.AuthToken{
		AccessToken:      token,
		ExpiresAt:        expiresAt,
		TokenType:        "Bearer",
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// GetByID retrieves a user by their ID.
func (s *authService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("authService.GetByID: %w", err)
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("authService.GetByID: %w", err)
	}

	return user, nil
}

// List returns a paginated list of users matching the given filter.
func (s *authService) List(ctx context.Context, f inbound.ListFilter) ([]*domain.User, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("authService.List: %w", err)
	}

	// Clamp pagination.
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 10
	}
	if f.Limit > 100 {
		f.Limit = 100
	}

	users, total, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("authService.List: %w", err)
	}

	return users, total, nil
}

// Update modifies an existing user.
func (s *authService) Update(ctx context.Context, id string, in inbound.UpdateInput) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("authService.Update: %w", err)
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("authService.Update: %w", err)
	}

	if in.Name != nil {
		if err := user.Rename(*in.Name); err != nil {
			return nil, fmt.Errorf("authService.Update: %w", apperrors.BadRequest(err.Error(), err))
		}
	}

	user.Touch(s.clock.Now())

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("authService.Update: %w", err)
	}

	s.log.Info("user updated", zap.String("user_id", user.ID))

	return user, nil
}

// Delete removes a user by ID.
func (s *authService) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("authService.Delete: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("authService.Delete: %w", err)
	}

	s.log.Info("user deleted", zap.String("user_id", id))

	return nil
}

// validatePassword enforces password policy: min 10 chars, upper, lower, digit, symbol.
func validatePassword(password string) error {
	if len(password) < 10 {
		return fmt.Errorf("%w: minimum 10 characters required", domain.ErrWeakPassword)
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSymbol = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: must contain at least one uppercase letter", domain.ErrWeakPassword)
	}
	if !hasLower {
		return fmt.Errorf("%w: must contain at least one lowercase letter", domain.ErrWeakPassword)
	}
	if !hasDigit {
		return fmt.Errorf("%w: must contain at least one digit", domain.ErrWeakPassword)
	}
	if !hasSymbol {
		return fmt.Errorf("%w: must contain at least one symbol", domain.ErrWeakPassword)
	}

	return nil
}

// RefreshToken validates a refresh token and issues a new access + refresh token pair.
func (s *authService) RefreshToken(ctx context.Context, in inbound.RefreshTokenInput) (*inbound.AuthToken, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}

	userID, role, err := s.tokens.VerifyRefresh(ctx, in.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, fmt.Errorf("authService.RefreshToken: %w", domain.ErrInvalidCredentials)
		}
		s.log.Error("unexpected error finding user for refresh", zap.Error(err))
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}

	if !user.Active {
		return nil, fmt.Errorf("authService.RefreshToken: %w", domain.ErrUserInactive)
	}

	// Re-read role from DB in case it changed since token was issued.
	_ = role

	token, expiresAt, err := s.tokens.Issue(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.tokens.IssueRefresh(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("authService.RefreshToken: %w", err)
	}

	s.log.Info("token refreshed", zap.String("user_id", user.ID))

	return &inbound.AuthToken{
		AccessToken:      token,
		ExpiresAt:        expiresAt,
		TokenType:        "Bearer",
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

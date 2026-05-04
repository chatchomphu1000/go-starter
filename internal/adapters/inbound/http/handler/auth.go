package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/dto"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// AuthHandler handles user-related HTTP requests.
type AuthHandler struct {
	svc inbound.AuthService
	log logger.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc inbound.AuthService, log logger.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Register handles POST /api/v1/auth/register
// @Summary      Register a new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Registration data"
// @Success      201   {object}  dto.UserResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.svc.Register(c.Request().Context(), inbound.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}

// Login handles POST /api/v1/auth/login
// @Summary      Login
// @Description  Authenticates a user and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  dto.LoginResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	token, err := h.svc.Login(c.Request().Context(), inbound.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:      token.AccessToken,
		TokenType:        token.TokenType,
		ExpiresAt:        token.ExpiresAt,
		RefreshToken:     token.RefreshToken,
		RefreshExpiresAt: token.RefreshExpiresAt,
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
// @Summary      Refresh JWT token
// @Description  Refreshes an access token using a valid refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshTokenRequest  true  "Refresh token"
// @Success      200   {object}  dto.LoginResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	token, err := h.svc.RefreshToken(c.Request().Context(), inbound.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken:      token.AccessToken,
		TokenType:        token.TokenType,
		ExpiresAt:        token.ExpiresAt,
		RefreshToken:     token.RefreshToken,
		RefreshExpiresAt: token.RefreshExpiresAt,
	})
}

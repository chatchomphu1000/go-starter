package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
)

// Auth returns middleware that verifies JWT tokens and injects user_id and role into the context.
func Auth(tokens outbound.TokenIssuer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				c.Response().Header().Set("WWW-Authenticate", "Bearer")
				return apperrors.Unauthorized("missing authorization header", nil)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				c.Response().Header().Set("WWW-Authenticate", "Bearer")
				return apperrors.Unauthorized("invalid authorization format", nil)
			}

			token := parts[1]
			userID, role, err := tokens.Verify(c.Request().Context(), token)
			if err != nil {
				c.Response().Header().Set("WWW-Authenticate", "Bearer")
				return err
			}

			c.Set("user_id", userID)
			c.Set("role", role.String())

			return next(c)
		}
	}
}

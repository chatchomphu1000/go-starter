package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
)

// RequireRole returns a middleware that enforces one of the given roles.
// The auth middleware must run first to populate "role" in the context.
func RequireRole(roles ...domain.Role) echo.MiddlewareFunc {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleStr, ok := c.Get("role").(string)
			if !ok || roleStr == "" {
				return echo.NewHTTPError(http.StatusUnauthorized,
					apperrors.Unauthorized("role not found in context", nil))
			}

			role := domain.Role(roleStr)
			if _, permitted := allowed[role]; !permitted {
				return echo.NewHTTPError(http.StatusForbidden,
					apperrors.Forbidden("insufficient role", nil))
			}

			return next(c)
		}
	}
}

package http

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// NewErrorHandler returns an Echo HTTPErrorHandler that maps errors to structured JSON responses.
func NewErrorHandler(log logger.Logger, isProduction bool) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		requestID := c.Response().Header().Get("X-Request-ID")

		// Handle domain errors first by mapping them to AppErrors.
		mapped := mapDomainError(err)

		var appErr *apperrors.AppError
		if errors.As(mapped, &appErr) {
			status, resp := apperrors.HTTPStatus(appErr, isProduction)
			resp.RequestID = requestID
			if status >= 500 {
				log.Error("internal error", zap.Error(err), zap.String("request_id", requestID))
			}
			if jsonErr := c.JSON(status, resp); jsonErr != nil {
				log.Error("failed to write error response", zap.Error(jsonErr))
			}
			return
		}

		// Handle validation errors from go-playground/validator.
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			details := make([]string, 0, len(validationErrs))
			for _, fe := range validationErrs {
				details = append(details, fe.Field()+": "+fe.Tag())
			}
			resp := apperrors.ErrorResponse{
				Code:      apperrors.CodeValidationFailed,
				Message:   "validation failed",
				Details:   details,
				RequestID: requestID,
			}
			if jsonErr := c.JSON(http.StatusBadRequest, resp); jsonErr != nil {
				log.Error("failed to write error response", zap.Error(jsonErr))
			}
			return
		}

		// Handle echo.HTTPError (e.g., from bind/validate).
		var he *echo.HTTPError
		if errors.As(err, &he) {
			// Try to unwrap internal validator errors.
			if he.Internal != nil {
				var ve validator.ValidationErrors
				if errors.As(he.Internal, &ve) {
					details := make([]string, 0, len(ve))
					for _, fe := range ve {
						details = append(details, fe.Field()+": "+fe.Tag())
					}
					resp := apperrors.ErrorResponse{
						Code:      apperrors.CodeValidationFailed,
						Message:   "validation failed",
						Details:   details,
						RequestID: requestID,
					}
					if jsonErr := c.JSON(http.StatusBadRequest, resp); jsonErr != nil {
						log.Error("failed to write error response", zap.Error(jsonErr))
					}
					return
				}
			}

			msg := ""
			if m, ok := he.Message.(string); ok {
				msg = m
			}
			resp := apperrors.ErrorResponse{
				Code:      apperrors.CodeBadRequest,
				Message:   msg,
				RequestID: requestID,
			}
			if he.Code >= 500 {
				resp.Code = apperrors.CodeInternal
				if isProduction {
					resp.Message = "internal server error"
				}
				log.Error("internal error", zap.Error(err), zap.String("request_id", requestID))
			}
			if jsonErr := c.JSON(he.Code, resp); jsonErr != nil {
				log.Error("failed to write error response", zap.Error(jsonErr))
			}
			return
		}

		// Unknown error — 500.
		log.Error("unhandled error", zap.Error(err), zap.String("request_id", requestID))
		resp := apperrors.ErrorResponse{
			Code:      apperrors.CodeInternal,
			Message:   "internal server error",
			RequestID: requestID,
		}
		if jsonErr := c.JSON(http.StatusInternalServerError, resp); jsonErr != nil {
			log.Error("failed to write error response", zap.Error(jsonErr))
		}
	}
}

// mapDomainError maps domain sentinel errors to apperrors.AppError.
func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return apperrors.NotFound("user not found", err)
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return apperrors.Conflict("email already exists", err)
	case errors.Is(err, domain.ErrInvalidCredentials):
		return apperrors.Unauthorized("invalid credentials", err)
	case errors.Is(err, domain.ErrInvalidEmail):
		return apperrors.BadRequest("invalid email", err)
	case errors.Is(err, domain.ErrInvalidRole):
		return apperrors.BadRequest("invalid role", err)
	case errors.Is(err, domain.ErrWeakPassword):
		return apperrors.BadRequest("password does not meet requirements", err)
	case errors.Is(err, domain.ErrUserInactive):
		return apperrors.Forbidden("user account is inactive", err)
	default:
		return err
	}
}

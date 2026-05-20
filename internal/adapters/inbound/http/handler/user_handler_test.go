package handler_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/handler"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
)

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestUserHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setup      func(svc *mocks.MockUserService)
		wantStatus int
		wantInBody string
	}{
		{
			name: "success_200",
			id:   "user-1",
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().GetByID(mock.Anything, "user-1").Return(validDomainUser(t), nil)
			},
			wantStatus: http.StatusOK,
			wantInBody: `"id":"user-1"`,
		},
		{
			name: "not_found_404",
			id:   "missing",
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().GetByID(mock.Anything, "missing").Return(nil, domain.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "response_no_hashed_password",
			id:   "user-1",
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().GetByID(mock.Anything, "user-1").Return(validDomainUser(t), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockUserService)
			tc.setup(svc)

			h := handler.NewUserHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodGet, "/api/v1/users/"+tc.id, "")
			c.SetParamNames("id")
			c.SetParamValues(tc.id)
			call(e, c, h.GetByID)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantInBody != "" {
				assert.Contains(t, rec.Body.String(), tc.wantInBody)
			}
			assert.NotContains(t, rec.Body.String(), "hashed_password")
			svc.AssertExpectations(t)
		})
	}
}

// ─── List ────────────────────────────────────────────────────────────────────

func TestUserHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setup      func(svc *mocks.MockUserService)
		wantStatus int
	}{
		{
			name:  "success_200",
			query: "?page=1&limit=10",
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]*domain.User{validDomainUser(t)}, int64(1), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "invalid_role_param_propagates_error",
			query: "?role=superuser",
			setup: func(_ *mocks.MockUserService) {},
			// ParseRole returns ErrInvalidRole → mapped to 400 by error handler
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockUserService)
			tc.setup(svc)

			h := handler.NewUserHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodGet, "/api/v1/users"+tc.query, "")
			call(e, c, h.List)

			assert.Equal(t, tc.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func TestUserHandler_Update(t *testing.T) {
	newName := "Bob"
	_ = newName

	tests := []struct {
		name       string
		id         string
		body       string
		setup      func(svc *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "success_200",
			id:   "user-1",
			body: `{"name":"Bob"}`,
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().Update(mock.Anything, "user-1", mock.Anything).
					Return(validDomainUser(t), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid_json_400",
			id:         "user-1",
			body:       `bad-json`,
			setup:      func(_ *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "name_too_short_400",
			id:   "user-1",
			body: `{"name":"A"}`,
			setup: func(_ *mocks.MockUserService) {
				// UpdateInput.Name pointer with short name passes bind but fails validation
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not_found_404",
			id:   "missing",
			body: `{"name":"Bob"}`,
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().Update(mock.Anything, "missing", mock.Anything).
					Return(nil, domain.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockUserService)
			tc.setup(svc)

			h := handler.NewUserHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodPut, "/api/v1/users/"+tc.id, tc.body)
			c.SetParamNames("id")
			c.SetParamValues(tc.id)
			call(e, c, h.Update)

			assert.Equal(t, tc.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestUserHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		setup      func(svc *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "success_204",
			id:   "user-1",
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().Delete(mock.Anything, "user-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not_found_404",
			id:   "missing",
			setup: func(svc *mocks.MockUserService) {
				svc.EXPECT().Delete(mock.Anything, "missing").Return(domain.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := makeEcho()
			svc := new(mocks.MockUserService)
			tc.setup(svc)

			h := handler.NewUserHandler(svc, nopLogger())
			c, rec := newRequest(t, e, http.MethodDelete, "/api/v1/users/"+tc.id, "")
			c.SetParamNames("id")
			c.SetParamValues(tc.id)
			call(e, c, h.Delete)

			assert.Equal(t, tc.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

// ─── UserService interface stub ───────────────────────────────────────────────

// Ensure MockUserService satisfies inbound.UserService at compile time.
var _ inbound.UserService = (*mocks.MockUserService)(nil)

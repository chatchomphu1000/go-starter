package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/dto"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	svc inbound.UserService
	log logger.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc inbound.UserService, log logger.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// GetByID handles GET /api/v1/users/:id
// @Summary      Get user by ID
// @Description  Returns a single user by their ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  dto.UserResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c echo.Context) error {
	id := c.Param("id")

	user, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// List handles GET /api/v1/users
// @Summary      List users
// @Description  Returns a paginated list of users
// @Tags         users
// @Produce      json
// @Param        page       query     int     false  "Page number"      default(1)
// @Param        limit      query     int     false  "Items per page"   default(10)
// @Param        sort_by    query     string  false  "Sort field"       Enums(created_at, name, email)
// @Param        sort_desc  query     bool    false  "Sort descending"
// @Param        role       query     string  false  "Filter by role"   Enums(admin, user)
// @Param        active     query     bool    false  "Filter by active"
// @Param        search     query     string  false  "Search by name/email"
// @Success      200        {object}  dto.ListResponse
// @Security     BearerAuth
// @Router       /api/v1/users [get]
func (h *UserHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	sortDesc, _ := strconv.ParseBool(c.QueryParam("sort_desc"))

	filter := inbound.ListFilter{
		Search:   c.QueryParam("search"),
		Page:     page,
		Limit:    limit,
		SortBy:   c.QueryParam("sort_by"),
		SortDesc: sortDesc,
	}

	if roleStr := c.QueryParam("role"); roleStr != "" {
		role, err := domain.ParseRole(roleStr)
		if err != nil {
			return err
		}
		filter.Role = &role
	}

	if activeStr := c.QueryParam("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			filter.Active = &active
		}
	}

	users, total, err := h.svc.List(c.Request().Context(), filter)
	if err != nil {
		return err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return c.JSON(http.StatusOK, dto.ToUserListResponse(users, total, filter.Page, filter.Limit))
}

// Update handles PUT /api/v1/users/:id
// @Summary      Update user
// @Description  Updates user information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "User ID"
// @Param        body  body      dto.UpdateRequest  true  "Update data"
// @Success      200   {object}  dto.UserResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) Update(c echo.Context) error {
	id := c.Param("id")

	var req dto.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.svc.Update(c.Request().Context(), id, inbound.UpdateInput{
		Name: req.Name,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// Delete handles DELETE /api/v1/users/:id
// @Summary      Delete user
// @Description  Deletes a user by their ID
// @Tags         users
// @Param        id   path  string  true  "User ID"
// @Success      204
// @Failure      404  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

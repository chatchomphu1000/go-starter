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

// NoticeHandler handles notice board HTTP requests.
type NoticeHandler struct {
	svc inbound.NoticeService
	log logger.Logger
}

// NewNoticeHandler creates a new NoticeHandler.
func NewNoticeHandler(svc inbound.NoticeService, log logger.Logger) *NoticeHandler {
	return &NoticeHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/notices
// @Summary      Post an announcement
// @Tags         notices
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateNoticeRequest  true  "Notice data"
// @Success      201   {object}  dto.NoticeResponse
// @Security     BearerAuth
// @Router       /api/v1/notices [post]
func (h *NoticeHandler) Create(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.CreateNoticeRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	n, err := h.svc.Create(c.Request().Context(), inbound.CreateNoticeInput{
		OwnerID:   ownerID,
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		Pinned:    req.Pinned,
		ExpiresAt: dto.ParseExpiresAt(req.ExpiresAt),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToNoticeResponse(n))
}

// GetByID handles GET /api/v1/notices/:id
// @Summary      Get a notice
// @Tags         notices
// @Produce      json
// @Param        id  path      string  true  "Notice ID"
// @Success      200 {object}  dto.NoticeResponse
// @Router       /api/v1/notices/{id} [get]
func (h *NoticeHandler) GetByID(c echo.Context) error {
	n, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToNoticeResponse(n))
}

// List handles GET /api/v1/notices
// @Summary      List notices
// @Tags         notices
// @Produce      json
// @Param        owner_id     query  string  false  "Filter by owner"
// @Param        type         query  string  false  "Filter by type"
// @Param        pinned_only  query  bool    false  "Show pinned only"
// @Param        active_only  query  bool    false  "Exclude expired"
// @Param        page         query  int     false  "Page"
// @Param        limit        query  int     false  "Limit"
// @Success      200          {object}  dto.NoticeListResponse
// @Router       /api/v1/notices [get]
func (h *NoticeHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	pinnedOnly, _ := strconv.ParseBool(c.QueryParam("pinned_only"))
	activeOnly, _ := strconv.ParseBool(c.QueryParam("active_only"))

	f := inbound.NoticeFilter{
		OwnerID:    c.QueryParam("owner_id"),
		PinnedOnly: pinnedOnly,
		ActiveOnly: activeOnly,
		Page:       page,
		Limit:      limit,
		SortDesc:   true,
	}
	if t := c.QueryParam("type"); t != "" {
		nt := domain.NoticeType(t)
		f.Type = &nt
	}

	notices, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	return c.JSON(http.StatusOK, dto.ToNoticeListResponse(notices, total, f.Page, f.Limit))
}

// Update handles PUT /api/v1/notices/:id
// @Summary      Update a notice
// @Tags         notices
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Notice ID"
// @Param        body  body      dto.UpdateNoticeRequest true  "Update data"
// @Success      200   {object}  dto.NoticeResponse
// @Security     BearerAuth
// @Router       /api/v1/notices/{id} [put]
func (h *NoticeHandler) Update(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.UpdateNoticeRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	n, err := h.svc.Update(c.Request().Context(), c.Param("id"), ownerID, inbound.UpdateNoticeInput{
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		Pinned:    req.Pinned,
		ExpiresAt: dto.ParseExpiresAt(req.ExpiresAt),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.ToNoticeResponse(n))
}

// Delete handles DELETE /api/v1/notices/:id
// @Summary      Delete a notice
// @Tags         notices
// @Param        id  path  string  true  "Notice ID"
// @Success      204
// @Security     BearerAuth
// @Router       /api/v1/notices/{id} [delete]
func (h *NoticeHandler) Delete(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), ownerID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

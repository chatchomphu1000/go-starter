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

// MaintenanceHandler handles maintenance ticket HTTP requests.
type MaintenanceHandler struct {
	svc inbound.MaintenanceService
	log logger.Logger
}

// NewMaintenanceHandler creates a new MaintenanceHandler.
func NewMaintenanceHandler(svc inbound.MaintenanceService, log logger.Logger) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/maintenance
// @Summary      Open a maintenance ticket
// @Tags         maintenance
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateTicketRequest  true  "Ticket data"
// @Success      201   {object}  dto.TicketResponse
// @Security     BearerAuth
// @Router       /api/v1/maintenance [post]
func (h *MaintenanceHandler) Create(c echo.Context) error {
	tenantID, _ := c.Get("user_id").(string)

	var req dto.CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ticket, err := h.svc.Create(c.Request().Context(), inbound.CreateTicketInput{
		RoomID:      req.RoomID,
		TenantID:    tenantID,
		OwnerID:     req.OwnerID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Photos:      req.Photos,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToTicketResponse(ticket))
}

// GetByID handles GET /api/v1/maintenance/:id
// @Summary      Get maintenance ticket
// @Tags         maintenance
// @Produce      json
// @Param        id  path      string  true  "Ticket ID"
// @Success      200 {object}  dto.TicketResponse
// @Security     BearerAuth
// @Router       /api/v1/maintenance/{id} [get]
func (h *MaintenanceHandler) GetByID(c echo.Context) error {
	t, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToTicketResponse(t))
}

// List handles GET /api/v1/maintenance
// @Summary      List maintenance tickets
// @Tags         maintenance
// @Produce      json
// @Param        room_id    query  string  false  "Filter by room"
// @Param        tenant_id  query  string  false  "Filter by tenant"
// @Param        owner_id   query  string  false  "Filter by owner"
// @Param        status     query  string  false  "Filter by status"
// @Param        priority   query  string  false  "Filter by priority"
// @Param        page       query  int     false  "Page"
// @Param        limit      query  int     false  "Limit"
// @Success      200        {object}  dto.TicketListResponse
// @Security     BearerAuth
// @Router       /api/v1/maintenance [get]
func (h *MaintenanceHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	f := inbound.TicketFilter{
		RoomID:   c.QueryParam("room_id"),
		TenantID: c.QueryParam("tenant_id"),
		OwnerID:  c.QueryParam("owner_id"),
		Page:     page,
		Limit:    limit,
		SortDesc: true,
	}
	if s := c.QueryParam("status"); s != "" {
		st := domain.TicketStatus(s)
		f.Status = &st
	}
	if p := c.QueryParam("priority"); p != "" {
		pr := domain.TicketPriority(p)
		f.Priority = &pr
	}

	tickets, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	return c.JSON(http.StatusOK, dto.ToTicketListResponse(tickets, total, f.Page, f.Limit))
}

// StartWork handles POST /api/v1/maintenance/:id/start
// @Summary      Start working on a ticket
// @Tags         maintenance
// @Param        id  path  string  true  "Ticket ID"
// @Success      200 {object}  dto.TicketResponse
// @Security     BearerAuth
// @Router       /api/v1/maintenance/{id}/start [post]
func (h *MaintenanceHandler) StartWork(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	t, err := h.svc.StartWork(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToTicketResponse(t))
}

// Resolve handles POST /api/v1/maintenance/:id/resolve
// @Summary      Resolve a maintenance ticket
// @Tags         maintenance
// @Accept       json
// @Param        id    path  string                      true  "Ticket ID"
// @Param        body  body  dto.ResolveTicketRequest   false "Resolution notes"
// @Success      200   {object}  dto.TicketResponse
// @Security     BearerAuth
// @Router       /api/v1/maintenance/{id}/resolve [post]
func (h *MaintenanceHandler) Resolve(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.ResolveTicketRequest
	_ = c.Bind(&req)

	t, err := h.svc.Resolve(c.Request().Context(), c.Param("id"), ownerID, req.Notes)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToTicketResponse(t))
}

// Close handles POST /api/v1/maintenance/:id/close
// @Summary      Close a resolved ticket (tenant confirmation)
// @Tags         maintenance
// @Param        id  path  string  true  "Ticket ID"
// @Success      200 {object}  dto.TicketResponse
// @Security     BearerAuth
// @Router       /api/v1/maintenance/{id}/close [post]
func (h *MaintenanceHandler) Close(c echo.Context) error {
	tenantID, _ := c.Get("user_id").(string)
	t, err := h.svc.Close(c.Request().Context(), c.Param("id"), tenantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToTicketResponse(t))
}

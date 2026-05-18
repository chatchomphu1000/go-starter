package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/dto"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// BookingHandler handles booking-related HTTP requests.
type BookingHandler struct {
	svc inbound.BookingService
	log logger.Logger
}

// NewBookingHandler creates a new BookingHandler.
func NewBookingHandler(svc inbound.BookingService, log logger.Logger) *BookingHandler {
	return &BookingHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/bookings
// @Summary      Create a booking
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateBookingRequest  true  "Booking data"
// @Success      201   {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings [post]
func (h *BookingHandler) Create(c echo.Context) error {
	tenantID, _ := c.Get("user_id").(string)

	var req dto.CreateBookingRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return echo.ErrBadRequest
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return echo.ErrBadRequest
	}

	booking, err := h.svc.Create(c.Request().Context(), inbound.CreateBookingInput{
		RoomID:    req.RoomID,
		TenantID:  tenantID,
		StartDate: start,
		EndDate:   end,
		Notes:     req.Notes,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToBookingResponse(booking))
}

// GetByID handles GET /api/v1/bookings/:id
// @Summary      Get booking by ID
// @Tags         bookings
// @Produce      json
// @Param        id  path      string  true  "Booking ID"
// @Success      200 {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings/{id} [get]
func (h *BookingHandler) GetByID(c echo.Context) error {
	b, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToBookingResponse(b))
}

// List handles GET /api/v1/bookings
// @Summary      List bookings
// @Tags         bookings
// @Produce      json
// @Param        owner_id   query  string  false  "Filter by owner"
// @Param        tenant_id  query  string  false  "Filter by tenant"
// @Param        room_id    query  string  false  "Filter by room"
// @Param        status     query  string  false  "Filter by status"
// @Param        page       query  int     false  "Page"   default(1)
// @Param        limit      query  int     false  "Limit"  default(20)
// @Success      200        {object}  dto.BookingListResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings [get]
func (h *BookingHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	f := inbound.BookingFilter{
		OwnerID:  c.QueryParam("owner_id"),
		TenantID: c.QueryParam("tenant_id"),
		RoomID:   c.QueryParam("room_id"),
		Page:     page,
		Limit:    limit,
		SortDesc: true,
	}
	if s := c.QueryParam("status"); s != "" {
		st := domain.BookingStatus(s)
		f.Status = &st
	}

	bookings, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	return c.JSON(http.StatusOK, dto.ToBookingListResponse(bookings, total, f.Page, f.Limit))
}

// Approve handles POST /api/v1/bookings/:id/approve
// @Summary      Approve a booking
// @Tags         bookings
// @Produce      json
// @Param        id  path  string  true  "Booking ID"
// @Success      200 {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings/{id}/approve [post]
func (h *BookingHandler) Approve(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	b, err := h.svc.Approve(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToBookingResponse(b))
}

// Reject handles POST /api/v1/bookings/:id/reject
// @Summary      Reject a booking
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id    path  string                      true  "Booking ID"
// @Param        body  body  dto.RejectBookingRequest   true  "Rejection reason"
// @Success      200   {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings/{id}/reject [post]
func (h *BookingHandler) Reject(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.RejectBookingRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	b, err := h.svc.Reject(c.Request().Context(), c.Param("id"), ownerID, req.Reason)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToBookingResponse(b))
}

// Cancel handles POST /api/v1/bookings/:id/cancel
// @Summary      Cancel a booking
// @Tags         bookings
// @Param        id  path  string  true  "Booking ID"
// @Success      200 {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings/{id}/cancel [post]
func (h *BookingHandler) Cancel(c echo.Context) error {
	requesterID, _ := c.Get("user_id").(string)
	b, err := h.svc.Cancel(c.Request().Context(), c.Param("id"), requesterID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToBookingResponse(b))
}

// Activate handles POST /api/v1/bookings/:id/activate
// @Summary      Activate a booking (tenant moved in)
// @Tags         bookings
// @Param        id  path  string  true  "Booking ID"
// @Success      200 {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings/{id}/activate [post]
func (h *BookingHandler) Activate(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	b, err := h.svc.Activate(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToBookingResponse(b))
}

// Complete handles POST /api/v1/bookings/:id/complete
// @Summary      Complete a booking (tenant moved out)
// @Tags         bookings
// @Param        id  path  string  true  "Booking ID"
// @Success      200 {object}  dto.BookingResponse
// @Security     BearerAuth
// @Router       /api/v1/bookings/{id}/complete [post]
func (h *BookingHandler) Complete(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	b, err := h.svc.Complete(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToBookingResponse(b))
}

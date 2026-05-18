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

// InvoiceHandler handles invoice-related HTTP requests.
type InvoiceHandler struct {
	svc inbound.InvoiceService
	log logger.Logger
}

// NewInvoiceHandler creates a new InvoiceHandler.
func NewInvoiceHandler(svc inbound.InvoiceService, log logger.Logger) *InvoiceHandler {
	return &InvoiceHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/invoices
// @Summary      Create an invoice
// @Tags         invoices
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateInvoiceRequest  true  "Invoice data"
// @Success      201   {object}  dto.InvoiceResponse
// @Security     BearerAuth
// @Router       /api/v1/invoices [post]
func (h *InvoiceHandler) Create(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.CreateInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return echo.ErrBadRequest
	}

	inv, err := h.svc.Create(c.Request().Context(), inbound.CreateInvoiceInput{
		BookingID: req.BookingID,
		TenantID:  req.TenantID,
		OwnerID:   ownerID,
		Items:     dto.ToInvoiceItemInput(req.Items),
		DueDate:   dueDate,
		Month:     req.Month,
		Year:      req.Year,
		Notes:     req.Notes,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToInvoiceResponse(inv))
}

// GetByID handles GET /api/v1/invoices/:id
// @Summary      Get invoice by ID
// @Tags         invoices
// @Produce      json
// @Param        id  path      string  true  "Invoice ID"
// @Success      200 {object}  dto.InvoiceResponse
// @Security     BearerAuth
// @Router       /api/v1/invoices/{id} [get]
func (h *InvoiceHandler) GetByID(c echo.Context) error {
	inv, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToInvoiceResponse(inv))
}

// List handles GET /api/v1/invoices
// @Summary      List invoices
// @Tags         invoices
// @Produce      json
// @Param        booking_id  query  string  false  "Filter by booking"
// @Param        tenant_id   query  string  false  "Filter by tenant"
// @Param        status      query  string  false  "Filter by status"
// @Param        month       query  int     false  "Filter by month (1-12)"
// @Param        year        query  int     false  "Filter by year"
// @Param        page        query  int     false  "Page"
// @Param        limit       query  int     false  "Limit"
// @Success      200         {object}  dto.InvoiceListResponse
// @Security     BearerAuth
// @Router       /api/v1/invoices [get]
func (h *InvoiceHandler) List(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	f := inbound.InvoiceFilter{
		BookingID: c.QueryParam("booking_id"),
		TenantID:  c.QueryParam("tenant_id"),
		OwnerID:   ownerID,
		Page:      page,
		Limit:     limit,
		SortDesc:  true,
	}
	if s := c.QueryParam("status"); s != "" {
		st := domain.InvoiceStatus(s)
		f.Status = &st
	}
	if m, err := strconv.Atoi(c.QueryParam("month")); err == nil && m > 0 {
		f.Month = &m
	}
	if y, err := strconv.Atoi(c.QueryParam("year")); err == nil && y > 0 {
		f.Year = &y
	}

	invoices, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	return c.JSON(http.StatusOK, dto.ToInvoiceListResponse(invoices, total, f.Page, f.Limit))
}

// Send handles POST /api/v1/invoices/:id/send
// @Summary      Send invoice to tenant
// @Tags         invoices
// @Param        id  path  string  true  "Invoice ID"
// @Success      200 {object}  dto.InvoiceResponse
// @Security     BearerAuth
// @Router       /api/v1/invoices/{id}/send [post]
func (h *InvoiceHandler) Send(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	inv, err := h.svc.Send(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToInvoiceResponse(inv))
}

// MarkPaid handles POST /api/v1/invoices/:id/pay
// @Summary      Mark invoice as paid
// @Tags         invoices
// @Param        id  path  string  true  "Invoice ID"
// @Success      200 {object}  dto.InvoiceResponse
// @Security     BearerAuth
// @Router       /api/v1/invoices/{id}/pay [post]
func (h *InvoiceHandler) MarkPaid(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	inv, err := h.svc.MarkPaid(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToInvoiceResponse(inv))
}

// Cancel handles POST /api/v1/invoices/:id/cancel
// @Summary      Cancel an invoice
// @Tags         invoices
// @Param        id  path  string  true  "Invoice ID"
// @Success      200 {object}  dto.InvoiceResponse
// @Security     BearerAuth
// @Router       /api/v1/invoices/{id}/cancel [post]
func (h *InvoiceHandler) Cancel(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	inv, err := h.svc.Cancel(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToInvoiceResponse(inv))
}

// Download handles GET /api/v1/invoices/:id/download
// @Summary      Download invoice as HTML (print-ready)
// @Tags         invoices
// @Produce      text/html
// @Param        id  path  string  true  "Invoice ID"
// @Success      200
// @Security     BearerAuth
// @Router       /api/v1/invoices/{id}/download [get]
func (h *InvoiceHandler) Download(c echo.Context) error {
	content, contentType, err := h.svc.GeneratePDF(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, contentType, content)
}

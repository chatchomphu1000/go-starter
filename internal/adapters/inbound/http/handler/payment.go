package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/dto"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// PaymentHandler handles payment-related HTTP requests.
type PaymentHandler struct {
	svc inbound.PaymentService
	log logger.Logger
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc inbound.PaymentService, log logger.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/payments
// @Summary      Create a payment record
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreatePaymentRequest  true  "Payment data"
// @Success      201   {object}  dto.PaymentResponse
// @Security     BearerAuth
// @Router       /api/v1/payments [post]
func (h *PaymentHandler) Create(c echo.Context) error {
	tenantID, _ := c.Get("user_id").(string)

	var req dto.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// OwnerID is resolved by the booking — service will look it up from booking.
	p, err := h.svc.Create(c.Request().Context(), inbound.CreatePaymentInput{
		BookingID:   req.BookingID,
		InvoiceID:   req.InvoiceID,
		TenantID:    tenantID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Method:      req.Method,
		Gateway:     req.Gateway,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToPaymentResponse(p))
}

// GetByID handles GET /api/v1/payments/:id
// @Summary      Get payment by ID
// @Tags         payments
// @Produce      json
// @Param        id  path      string  true  "Payment ID"
// @Success      200 {object}  dto.PaymentResponse
// @Security     BearerAuth
// @Router       /api/v1/payments/{id} [get]
func (h *PaymentHandler) GetByID(c echo.Context) error {
	p, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToPaymentResponse(p))
}

// List handles GET /api/v1/payments
// @Summary      List payments
// @Tags         payments
// @Produce      json
// @Param        booking_id  query  string  false  "Filter by booking"
// @Param        tenant_id   query  string  false  "Filter by tenant"
// @Param        owner_id    query  string  false  "Filter by owner"
// @Param        status      query  string  false  "Filter by status"
// @Param        page        query  int     false  "Page"
// @Param        limit       query  int     false  "Limit"
// @Success      200         {object}  dto.PaymentListResponse
// @Security     BearerAuth
// @Router       /api/v1/payments [get]
func (h *PaymentHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	f := inbound.PaymentFilter{
		BookingID: c.QueryParam("booking_id"),
		InvoiceID: c.QueryParam("invoice_id"),
		TenantID:  c.QueryParam("tenant_id"),
		OwnerID:   c.QueryParam("owner_id"),
		Page:      page,
		Limit:     limit,
		SortDesc:  true,
	}
	if s := c.QueryParam("status"); s != "" {
		st := domain.PaymentStatus(s)
		f.Status = &st
	}

	payments, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	return c.JSON(http.StatusOK, dto.ToPaymentListResponse(payments, total, f.Page, f.Limit))
}

// Webhook handles POST /api/v1/payments/webhook/:gateway
// @Summary      Payment gateway webhook
// @Tags         payments
// @Accept       json
// @Param        gateway  path  string  true  "Gateway name (stripe|omise|paypal)"
// @Success      200
// @Router       /api/v1/payments/webhook/{gateway} [post]
func (h *PaymentHandler) Webhook(c echo.Context) error {
	gateway := c.Param("gateway")

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.ErrBadRequest
	}

	var req dto.WebhookRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest
	}

	if err := h.svc.HandleWebhook(c.Request().Context(), inbound.WebhookInput{
		Gateway:     gateway,
		RawPayload:  string(body),
		GatewayTxID: req.GatewayTxID,
		PaymentID:   req.PaymentID,
		Status:      req.Status,
	}); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Refund handles POST /api/v1/payments/:id/refund
// @Summary      Refund a payment
// @Tags         payments
// @Param        id  path  string  true  "Payment ID"
// @Success      200 {object}  dto.PaymentResponse
// @Security     BearerAuth
// @Router       /api/v1/payments/{id}/refund [post]
func (h *PaymentHandler) Refund(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	p, err := h.svc.Refund(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToPaymentResponse(p))
}

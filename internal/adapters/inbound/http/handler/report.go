package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// ReportHandler handles analytics and reporting HTTP requests.
type ReportHandler struct {
	svc inbound.ReportService
	log logger.Logger
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(svc inbound.ReportService, log logger.Logger) *ReportHandler {
	return &ReportHandler{svc: svc, log: log}
}

// OwnerDashboard handles GET /api/v1/owners/:id/dashboard
// @Summary      Owner dashboard summary
// @Tags         reports
// @Produce      json
// @Param        id  path  string  true  "Owner ID"
// @Success      200 {object}  inbound.OwnerDashboard
// @Security     BearerAuth
// @Router       /api/v1/owners/{id}/dashboard [get]
func (h *ReportHandler) OwnerDashboard(c echo.Context) error {
	ownerID := c.Param("id")
	dashboard, err := h.svc.OwnerDashboard(c.Request().Context(), ownerID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dashboard)
}

// IncomeExpense handles GET /api/v1/reports/income
// @Summary      Income and expense report
// @Tags         reports
// @Produce      json
// @Param        month  query  int  false  "Month (1-12)"
// @Param        year   query  int  false  "Year"
// @Success      200    {object}  inbound.IncomeExpenseReport
// @Security     BearerAuth
// @Router       /api/v1/reports/income [get]
func (h *ReportHandler) IncomeExpense(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	now := time.Now()

	month, _ := strconv.Atoi(c.QueryParam("month"))
	year, _ := strconv.Atoi(c.QueryParam("year"))
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 2020 {
		year = now.Year()
	}

	report, err := h.svc.IncomeExpense(c.Request().Context(), ownerID, month, year)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, report)
}

// TenantStatement handles GET /api/v1/tenants/:id/history
// @Summary      Tenant payment history
// @Tags         reports
// @Produce      json
// @Param        id     path   string  true  "Tenant ID"
// @Param        month  query  int     false  "Month"
// @Param        year   query  int     false  "Year"
// @Success      200    {object}  inbound.IncomeExpenseReport
// @Security     BearerAuth
// @Router       /api/v1/tenants/{id}/history [get]
func (h *ReportHandler) TenantStatement(c echo.Context) error {
	tenantID := c.Param("id")
	now := time.Now()

	month, _ := strconv.Atoi(c.QueryParam("month"))
	year, _ := strconv.Atoi(c.QueryParam("year"))
	if month < 1 || month > 12 {
		month = int(now.Month())
	}
	if year < 2020 {
		year = now.Year()
	}

	report, err := h.svc.TenantStatement(c.Request().Context(), tenantID, month, year)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, report)
}

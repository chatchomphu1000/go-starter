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

// RoomHandler handles room-related HTTP requests.
type RoomHandler struct {
	svc inbound.RoomService
	log logger.Logger
}

// NewRoomHandler creates a new RoomHandler.
func NewRoomHandler(svc inbound.RoomService, log logger.Logger) *RoomHandler {
	return &RoomHandler{svc: svc, log: log}
}

// Create handles POST /api/v1/rooms
// @Summary      Create a room
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateRoomRequest  true  "Room data"
// @Success      201   {object}  dto.RoomResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/rooms [post]
func (h *RoomHandler) Create(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.CreateRoomRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	room, err := h.svc.Create(c.Request().Context(), inbound.CreateRoomInput{
		OwnerID:     ownerID,
		Number:      req.Number,
		Name:        req.Name,
		Type:        req.Type,
		Floor:       req.Floor,
		SizeSqm:     req.SizeSqm,
		RentPrice:   req.RentPrice,
		Deposit:     req.Deposit,
		Amenities:   req.Amenities,
		Photos:      req.Photos,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToRoomResponse(room))
}

// GetByID handles GET /api/v1/rooms/:id
// @Summary      Get room by ID
// @Tags         rooms
// @Produce      json
// @Param        id   path      string  true  "Room ID"
// @Success      200  {object}  dto.RoomResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/rooms/{id} [get]
func (h *RoomHandler) GetByID(c echo.Context) error {
	room, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToRoomResponse(room))
}

// List handles GET /api/v1/rooms
// @Summary      List rooms
// @Tags         rooms
// @Produce      json
// @Param        owner_id    query     string  false  "Filter by owner ID"
// @Param        status      query     string  false  "Filter by status"
// @Param        type        query     string  false  "Filter by type"
// @Param        min_price   query     number  false  "Min rent price"
// @Param        max_price   query     number  false  "Max rent price"
// @Param        search      query     string  false  "Search by name/number"
// @Param        page        query     int     false  "Page"      default(1)
// @Param        limit       query     int     false  "Limit"     default(20)
// @Success      200         {object}  dto.RoomListResponse
// @Router       /api/v1/rooms [get]
func (h *RoomHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	f := inbound.RoomFilter{
		OwnerID:  c.QueryParam("owner_id"),
		Search:   c.QueryParam("search"),
		Page:     page,
		Limit:    limit,
		SortDesc: c.QueryParam("sort_desc") == "true",
	}

	if s := c.QueryParam("status"); s != "" {
		st := domain.RoomStatus(s)
		f.Status = &st
	}
	if t := c.QueryParam("type"); t != "" {
		rt := domain.RoomType(t)
		f.Type = &rt
	}
	if v, err := strconv.ParseFloat(c.QueryParam("min_price"), 64); err == nil {
		f.MinPrice = &v
	}
	if v, err := strconv.ParseFloat(c.QueryParam("max_price"), 64); err == nil {
		f.MaxPrice = &v
	}

	rooms, total, err := h.svc.List(c.Request().Context(), f)
	if err != nil {
		return err
	}

	page = f.Page
	limit = f.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return c.JSON(http.StatusOK, dto.ToRoomListResponse(rooms, total, page, limit))
}

// Update handles PUT /api/v1/rooms/:id
// @Summary      Update a room
// @Tags         rooms
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Room ID"
// @Param        body  body      dto.UpdateRoomRequest  true  "Update data"
// @Success      200   {object}  dto.RoomResponse
// @Security     BearerAuth
// @Router       /api/v1/rooms/{id} [put]
func (h *RoomHandler) Update(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)

	var req dto.UpdateRoomRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	room, err := h.svc.Update(c.Request().Context(), c.Param("id"), ownerID, inbound.UpdateRoomInput{
		Name:        req.Name,
		Type:        req.Type,
		Floor:       req.Floor,
		SizeSqm:     req.SizeSqm,
		RentPrice:   req.RentPrice,
		Deposit:     req.Deposit,
		Amenities:   req.Amenities,
		Photos:      req.Photos,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToRoomResponse(room))
}

// Delete handles DELETE /api/v1/rooms/:id
// @Summary      Delete a room
// @Tags         rooms
// @Param        id   path  string  true  "Room ID"
// @Success      204
// @Security     BearerAuth
// @Router       /api/v1/rooms/{id} [delete]
func (h *RoomHandler) Delete(c echo.Context) error {
	ownerID, _ := c.Get("user_id").(string)
	if err := h.svc.Delete(c.Request().Context(), c.Param("id"), ownerID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

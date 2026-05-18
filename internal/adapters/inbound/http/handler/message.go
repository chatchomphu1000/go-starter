package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/dto"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// MessageHandler handles chat messaging HTTP requests.
type MessageHandler struct {
	svc inbound.MessageService
	log logger.Logger
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(svc inbound.MessageService, log logger.Logger) *MessageHandler {
	return &MessageHandler{svc: svc, log: log}
}

// Send handles POST /api/v1/messages
// @Summary      Send a message
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SendMessageRequest  true  "Message data"
// @Success      201   {object}  dto.MessageResponse
// @Security     BearerAuth
// @Router       /api/v1/messages [post]
func (h *MessageHandler) Send(c echo.Context) error {
	senderID, _ := c.Get("user_id").(string)

	var req dto.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	msg, err := h.svc.Send(c.Request().Context(), inbound.SendMessageInput{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.ToMessageResponse(msg))
}

// ListThreads handles GET /api/v1/messages/threads
// @Summary      List conversation threads
// @Tags         messages
// @Produce      json
// @Param        page   query  int  false  "Page"
// @Param        limit  query  int  false  "Limit"
// @Success      200    {object}  dto.ThreadListResponse
// @Security     BearerAuth
// @Router       /api/v1/messages/threads [get]
func (h *MessageHandler) ListThreads(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	threads, total, err := h.svc.ListThreads(c.Request().Context(), userID, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToThreadListResponse(threads, total, page, limit))
}

// ListMessages handles GET /api/v1/messages/threads/:threadId
// @Summary      Get messages in a thread
// @Tags         messages
// @Produce      json
// @Param        threadId  path  string  true  "Thread ID"
// @Param        page      query  int    false  "Page"
// @Param        limit     query  int    false  "Limit"
// @Success      200       {object}  dto.MessageListResponse
// @Security     BearerAuth
// @Router       /api/v1/messages/threads/{threadId} [get]
func (h *MessageHandler) ListMessages(c echo.Context) error {
	threadID := c.Param("threadId")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	msgs, total, err := h.svc.ListMessages(c.Request().Context(), inbound.MessageFilter{
		ThreadID: threadID,
		Page:     page,
		Limit:    limit,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.ToMessageListResponse(msgs, total, page, limit))
}

// MarkRead handles POST /api/v1/messages/threads/:threadId/read
// @Summary      Mark thread messages as read
// @Tags         messages
// @Param        threadId  path  string  true  "Thread ID"
// @Success      204
// @Security     BearerAuth
// @Router       /api/v1/messages/threads/{threadId}/read [post]
func (h *MessageHandler) MarkRead(c echo.Context) error {
	userID, _ := c.Get("user_id").(string)
	if err := h.svc.MarkRead(c.Request().Context(), c.Param("threadId"), userID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

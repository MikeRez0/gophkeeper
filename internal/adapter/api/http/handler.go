package http

import (
	"net/http"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var errorStatusMap = map[error]int{
	domain.ErrInternal:        http.StatusInternalServerError,
	domain.ErrDataNotFound:    http.StatusNotFound,
	domain.ErrConflictingData: http.StatusConflict,

	domain.ErrInvalidCredentials:         http.StatusUnauthorized,
	domain.ErrUnauthorized:               http.StatusUnauthorized,
	domain.ErrEmptyAuthorizationHeader:   http.StatusUnauthorized,
	domain.ErrInvalidAuthorizationHeader: http.StatusUnauthorized,
	domain.ErrInvalidAuthorizationType:   http.StatusUnauthorized,
	domain.ErrInvalidToken:               http.StatusUnauthorized,
	domain.ErrExpiredToken:               http.StatusUnauthorized,
	domain.ErrForbidden:                  http.StatusForbidden,

	domain.ErrNoUpdatedData: http.StatusBadRequest,
	domain.ErrBadRequest:    http.StatusBadRequest,
}

// Handler - common logic for processing request.
type Handler struct {
	logger *zap.Logger
}

// NewHandler - creates new Handler object.
func NewHandler(logger *zap.Logger) *Handler {
	return &Handler{logger: logger}
}

// handleValidationError sends an error response for some specific request validation error.
func (h *Handler) handleValidationError(ctx *gin.Context, err error) {
	_ = ctx.AbortWithError(http.StatusBadRequest, err)
}

// handleAbort sends an error response and aborts the request with the specified status code and error message.
func (h *Handler) handleAbort(ctx *gin.Context, err error) {
	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
		h.logger.Error("aborting request", zap.Error(err))
	}
	_ = ctx.AbortWithError(statusCode, err)
}

func (h *Handler) handleError(ctx *gin.Context, err error) {
	statusCode, ok := errorStatusMap[err]
	if !ok {
		statusCode = http.StatusInternalServerError
		h.logger.Error("error processing request", zap.Error(err))
	}
	ctx.Status(statusCode)
}

// handleSuccess sends a success response with the specified status code and optional data.
func (h *Handler) handleSuccessWithStatus(ctx *gin.Context, data any, status int) {
	if data != nil {
		ctx.JSON(status, data)
	} else {
		ctx.Status(status)
	}
}

// handleSuccess sends a success response with 200 status code and optional data.
func (h *Handler) handleSuccess(ctx *gin.Context, data any) {
	h.handleSuccessWithStatus(ctx, data, http.StatusOK)
}

package http

import (
	"strings"
	"time"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const authHeaderKey = "Authorization"
const authType = "Bearer"
const userPayloadKey = "user_payload"

// authCheck - checks and validates authorization header.
func authCheck(tokenService port.TokenService, logger *zap.Logger) gin.HandlerFunc {
	authHandler := &Handler{logger: logger}
	return func(ctx *gin.Context) {
		header := ctx.Request.Header.Get(authHeaderKey)
		if len(header) == 0 {
			authHandler.handleAbort(ctx, domain.ErrEmptyAuthorizationHeader)
			return
		}

		words := strings.Split(header, " ")
		if len(words) != 2 {
			authHandler.handleAbort(ctx, domain.ErrInvalidAuthorizationHeader)
			return
		}
		if words[0] != authType {
			authHandler.handleAbort(ctx, domain.ErrInvalidAuthorizationType)
			return
		}
		token := words[1]
		payload, err := tokenService.VerifyToken(token)
		if err != nil {
			authHandler.handleAbort(ctx, domain.ErrInvalidToken)
			return
		}

		ctx.Set(userPayloadKey, payload)

		ctx.Next()
	}
}

// getAuthPayload - reads payload of auth token.
func getAuthPayload(ctx *gin.Context) *port.TokenPayload {
	if t, ok := ctx.MustGet(userPayloadKey).(*port.TokenPayload); ok {
		return t
	} else {
		return nil
	}
}

// logRequest - write request to log.
func logRequest(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		c.Next()

		log.Info("IncHTTP>",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.String("duration", time.Since(t).String()))
	}
}

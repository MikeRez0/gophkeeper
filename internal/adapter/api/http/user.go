package http

import (
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler - handler for user requests.
type UserHandler struct {
	Handler
	service port.IUserService
}

type userRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// NewUserHandler - creates new user handler.
func NewUserHandler(service port.IUserService, logger *zap.Logger) (*UserHandler, error) {
	return &UserHandler{
		Handler: Handler{logger: logger},
		service: service}, nil
}

// RegisterUser - create new user, authenticate and return token.
func (uh *UserHandler) RegisterUser(ctx *gin.Context) {
	userReq := userRequest{}
	err := ctx.ShouldBindBodyWithJSON(&userReq)
	if err != nil {
		uh.handleValidationError(ctx, err)
		return
	}

	user := &domain.User{
		Login:    userReq.Login,
		Password: userReq.Password,
	}

	_, err = uh.service.RegisterUser(ctx, user)
	if err != nil {
		uh.handleError(ctx, err)
		return
	}

	// Token return
	uh.LoginUser(ctx)
}

// LoginUser - authenticate user and return token.
func (uh *UserHandler) LoginUser(ctx *gin.Context) {
	userReq := userRequest{}
	err := ctx.ShouldBindBodyWithJSON(&userReq)
	if err != nil {
		uh.handleValidationError(ctx, err)
		return
	}

	token, err := uh.service.LoginUser(ctx, userReq.Login, userReq.Password)
	if err != nil {
		uh.handleError(ctx, err)
		return
	}

	ctx.Header(authHeaderKey, authType+" "+token)
	uh.handleSuccess(ctx, struct {
		Token string `json:"token"`
	}{Token: token})
}

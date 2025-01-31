package http

import (
	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	*gin.Engine
}

func NewRouter(
	conf *config.HTTP,
	tokenService port.TokenService,
	userHandler *UserHandler,
	log *zap.Logger) (*Router, error) {
	router := gin.New()

	router.Use(logRequest(log))

	router.Use(gzip.Gzip(gzip.BestSpeed))

	api := router.Group("/api")
	{
		user := api.Group("/user")
		{
			user.POST("/register", userHandler.RegisterUser)
			user.POST("/login", userHandler.LoginUser)

			orders := user.Group("/orders")
			{
				orders.Use(authCheck(tokenService, log))
				// orders.POST("", orderHandler.CreateOrder)
				// orders.GET("", orderHandler.ListOrdersByUser)
			}
		}
	}

	return &Router{router}, nil
}

// Serve starts the HTTP server.
func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}

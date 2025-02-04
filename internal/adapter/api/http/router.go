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
	keychainHandler *KeychainHandler,
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
		}

		keychain := api.Group("/keychain")
		{
			keychain.Use(authCheck(tokenService, log))

			keychain.GET("", keychainHandler.GetKeychainList)
			keychain.POST("/:"+cKeychainParamName, keychainHandler.SaveKeychain)
			keychain.GET("/:"+cKeychainParamName, keychainHandler.GetKeychain)
			items := keychain.Group("/:" + cKeychainParamName)
			{
				items.GET("/item", keychainHandler.ListKeychainItems)
				items.GET("/item/:"+cKeychainItemParamName, keychainHandler.GetKeychainItem)
				items.POST("/item/:"+cKeychainItemParamName, keychainHandler.SaveKeychainItem)
			}
		}
	}

	return &Router{router}, nil
}

// Serve starts the HTTP server.
func (r *Router) Serve(listenAddr string) error {
	return r.Run(listenAddr)
}

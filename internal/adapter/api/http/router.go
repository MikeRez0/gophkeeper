// Package http provides implementation of GophKeeper's HTTP API-server router.
package http

import (
	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router - main object for handling API requests.
type Router struct {
	*gin.Engine
}

// NewRouter creates new Router with user, keychain handlers.
//
// Also needs token service for authenticate requests.
func NewRouter(
	conf *config.HTTP, // HTTP-configuration
	tokenService port.TokenService, // Service for generate, validate user-token
	userHandler *UserHandler, // Handler for user requests
	keychainHandler *KeychainHandler, // Handler for keychain requests
	log *zap.Logger, // Log-object
) (*Router, error) {
	router := gin.New()

	router.Use(logRequest(log.WithOptions(zap.WithCaller(false))))

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
				items.POST("/sync", keychainHandler.Sync)
				items.GET("/item", keychainHandler.ListKeychainItems)
				items.GET("/item/:"+cKeychainItemParamName, keychainHandler.GetKeychainItem)
				items.POST("/item/:"+cKeychainItemParamName, keychainHandler.SaveKeychainItem)
			}
		}
	}

	return &Router{router}, nil
}

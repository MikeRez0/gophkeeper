package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/MikeRez0/gophkeeper/internal/adapter/api/http"
	"github.com/MikeRez0/gophkeeper/internal/adapter/auth"
	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/repository"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
)

var buildVersion string
var buildDate string
var buildCommit string

const cBuildInfoTemplate = `Server GophKeeper
Build version: %s
Build date: %s
Build commit: %s
`

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf(cBuildInfoTemplate, buildVersion, buildDate, buildCommit)

	conf, err := config.NewConfigServer()
	if err != nil {
		fmt.Printf("config error:%s", err)
		return
	}

	log := logger.NewLogger(conf.App)
	if log == nil {
		fmt.Printf("error creating log")
		return
	}
	defer func() {
		err := log.Sync()
		if err != nil {
			fmt.Printf("log error: %s", err)
		}
	}()

	ctx := context.Background()

	db, err := storage.NewDBStorage(ctx, conf.Database)
	if err != nil {
		log.Error("database error", zap.Error(err))
		return
	}
	err = db.RunMigrations()
	if err != nil {
		log.Error("database migration error", zap.Error(err))
		return
	}

	userRepo, err := repository.NewUserRepository(db)
	if err != nil {
		log.Error("order repo creating error", zap.Error(err))
		return
	}
	tokenService, err := auth.New()
	if err != nil {
		log.Error("token service creating error", zap.Error(err))
		return
	}

	userSrv, err := service.NewUserService(userRepo, tokenService, log.Named("UserService"))
	if err != nil {
		log.Error("order service creating error", zap.Error(err))
		return
	}

	userHandler, err := http.NewUserHandler(userSrv, log.Named("User handler"))
	if err != nil {
		log.Error("user handler creating error", zap.Error(err))
		return
	}

	kcRepo, err := repository.NewKeychainRepository(db)
	if err != nil {
		log.Error("order repo creating error", zap.Error(err))
		return
	}
	keychainSrv, err := service.NewKeychainService(kcRepo, log.Named("KeychainService"))

	kHandler, err := http.NewKeychainHandler(keychainSrv, log.Named("Keychain handler"))

	r, err := http.NewRouter(conf.HTTP, tokenService, userHandler, kHandler, log.Named("Router"))
	if err != nil {
		log.Error("router creating error", zap.Error(err))
		return
	}

	err = r.Serve(conf.HTTP.HostString)
	if err != nil {
		log.Error("router serve error", zap.Error(err))
		return
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	apiHttp "github.com/MikeRez0/gophkeeper/internal/adapter/api/http"
	"github.com/MikeRez0/gophkeeper/internal/adapter/auth"
	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/pgsql"
	sql "github.com/MikeRez0/gophkeeper/internal/adapter/storage/sqlite"
	"github.com/MikeRez0/gophkeeper/internal/core/port"
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

	l := logger.NewLogger(conf.App)
	if l == nil {
		fmt.Printf("error creating log")
		return
	}
	defer func() {
		err := l.Sync()
		if err != nil {
			fmt.Printf("log error: %s", err)
		}
	}()

	ctx := context.Background()

	var userRepo port.IUserRepository
	var keychainRepo port.IKeychainRepository

	if conf.Database.Driver == "postgresql" {
		db, err := pgsql.NewDBStorage(ctx, conf.Database)
		if err != nil {
			l.Error("postgresql database error", zap.Error(err))
			return
		}
		err = db.RunMigrations()
		if err != nil {
			l.Error("postgresql database migration error", zap.Error(err))
			return
		}

		userRepo, err = pgsql.NewUserRepository(db)
		if err != nil {
			l.Error("user repo (postgresql) creating error", zap.Error(err))
			return
		}
		keychainRepo, err = pgsql.NewKeychainPgRepository(db, l.Named("PgRepo"))
		if err != nil {
			l.Error("keychain repo (postgresql) creating error", zap.Error(err))
			return
		}
	} else {
		db, err := sql.NewDBStorage(ctx, conf.Database)
		if err != nil {
			l.Error("database error", zap.Error(err))
			return
		}
		err = db.RunMigrations()
		if err != nil {
			l.Error("database migration error", zap.Error(err))
			return
		}

		userRepo, err = sql.NewUserRepository(db, l)
		if err != nil {
			l.Error("user repo creating error", zap.Error(err))
			return
		}
		keychainRepo, err = sql.NewKeychainSqliteRepository(db, l.Named("SqlRepo"))
		if err != nil {
			l.Error("keychain repo creating error", zap.Error(err))
			return
		}
	}

	tokenService, err := auth.New()
	if err != nil {
		l.Error("token service creating error", zap.Error(err))
		return
	}

	userSrv, err := service.NewUserService(userRepo, tokenService, l.Named("UserService"))
	if err != nil {
		l.Error("order service creating error", zap.Error(err))
		return
	}

	userHandler, err := apiHttp.NewUserHandler(userSrv, l.Named("User handler"))
	if err != nil {
		l.Error("user handler creating error", zap.Error(err))
		return
	}

	keychainSrv, err := service.NewKeychainDataService(keychainRepo, l.Named("KeychainService"))
	if err != nil {
		l.Error("keychain service creating error", zap.Error(err))
		return
	}

	kHandler, err := apiHttp.NewKeychainHandler(keychainSrv, l.Named("Keychain handler"))
	if err != nil {
		l.Error("keychain handler creating error", zap.Error(err))
		return
	}

	r, err := apiHttp.NewRouter(conf.HTTP, tokenService, userHandler, kHandler, l.Named("Router"))
	if err != nil {
		l.Error("router creating error", zap.Error(err))
		return
	}

	server := &http.Server{
		Addr:    conf.HTTP.HostString,
		Handler: r.Handler(),
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-shutdown
		l.Info("Start gracefull shutdown...")

		err := server.Shutdown(context.Background())
		if err != nil {
			l.Error("error while shutdown", zap.Error(err))
		}
	}()

	if conf.HTTP.TLSCertFile != "" && conf.HTTP.TLSKeyFile != "" {
		err = server.ListenAndServeTLS(conf.HTTP.TLSCertFile, conf.HTTP.TLSKeyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("router serve error", zap.Error(err))
		return
	}

	l.Info("Server was shut down gracefully")
	_ = l.Sync()
}

package client

import (
	"context"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/sqlite"
	"github.com/MikeRez0/gophkeeper/internal/client/app"
	"github.com/MikeRez0/gophkeeper/internal/client/tui"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
	"go.uber.org/zap"
)

func Run() error {
	conf, err := config.NewConfigClient()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	log := logger.NewLogger(conf.App)

	ctx := context.Background()

	db, err := sqlite.NewDBStorage(ctx, conf.Database)
	if err != nil {
		log.Error("error creating database app", zap.Error(err))
		return err
	}
	err = db.RunMigrations()
	if err != nil {
		log.Error("error running database migration", zap.Error(err))
		return err
	}

	repo, err := sqlite.NewKeychainSqliteRepository(db, log.Named("repo"))
	if err != nil {
		log.Error("error creating keychain repo", zap.Error(err))
		return err
	}

	srv, err := service.NewKeychainDataService(repo, log.Named("service"))
	if err != nil {
		log.Error("error creating keychain service", zap.Error(err))
		return err
	}
	srv.SetLocalMode(true)

	app, err := app.NewApp(conf, srv, log)
	if err != nil {
		log.Error("error creating client app", zap.Error(err))
		return err
	}

	c, err := tui.NewUIController(app, log.Named("tui"))
	if err != nil {
		log.Error("error creating TUI", zap.Error(err))
		return err
	}

	err = c.Run()
	if err != nil {
		log.Error("error running app", zap.Error(err))
		return err
	}

	return nil
}

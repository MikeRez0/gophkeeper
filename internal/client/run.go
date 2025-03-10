package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage/sqlite"
	"github.com/MikeRez0/gophkeeper/internal/client/app"
	"github.com/MikeRez0/gophkeeper/internal/client/tui"
	"github.com/MikeRez0/gophkeeper/internal/core/service"
)

func BootstrapApp(conf *config.ConfigClient) (*app.ClientApp, error) {
	err := conf.Parse()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	log := logger.NewLogger(conf.App)
	if log == nil {
		return nil, errors.New("log not created")
	}

	ctx := context.Background()

	db, err := sqlite.NewDBStorage(ctx, conf.Database)
	if err != nil {
		return nil, fmt.Errorf("error creating database app: %w", err)
	}
	err = db.RunMigrations()
	if err != nil {
		return nil, fmt.Errorf("error running database migration: %w", err)
	}

	repo, err := sqlite.NewKeychainSqliteRepository(db, log.Named("repo"))
	if err != nil {
		return nil, fmt.Errorf("error creating keychain repo: %w", err)
	}

	srv, err := service.NewKeychainDataService(repo, log.Named("service"))
	if err != nil {
		return nil, fmt.Errorf("error creating keychain service: %w", err)
	}
	srv.SetLocalMode(true)

	a, err := app.NewApp(conf, srv, log)
	if err != nil {
		return nil, fmt.Errorf("error creating client app: %w", err)
	}

	return a, nil
}

func RunTUI(conf *config.ConfigClient) error {
	a, err := BootstrapApp(conf)
	if err != nil {
		return fmt.Errorf("bootstrap app error: %w", err)
	}

	c, err := tui.NewUIController(a, a.Log.Named("tui"))
	if err != nil {
		return fmt.Errorf("error creating TUI: %w", err)
	}

	err = c.Run()
	if err != nil {
		return fmt.Errorf("error running TUI app: %w", err)
	}

	return nil
}

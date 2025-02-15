package client

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/client/app"
	"github.com/MikeRez0/gophkeeper/internal/client/tui"
	"go.uber.org/zap"
)

func Run() error {
	conf, err := config.NewConfigClient()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	log := logger.NewLogger(conf.App)

	app, err := app.NewApp(conf, log)
	if err != nil {
		log.Error("error creating client app", zap.Error(err))
		return err
	}

	c, err := tui.NewUIController(app, log)
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

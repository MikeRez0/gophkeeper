package logger

import (
	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(conf *config.App) *zap.Logger {
	lvl, err := zap.ParseAtomicLevel(conf.LogLevel)
	if err != nil {
		zap.L().Error("error parsing log level", zap.Error(err))
		return nil
	}

	if conf.Mode == config.AppModeDevelop {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = lvl
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		var options []zap.Option
		if conf.LogLevel == "debug" {
			options = append(options, zap.AddStacktrace(zap.ErrorLevel))
		} else {
			options = append(options, zap.AddStacktrace(zap.FatalLevel))
		}

		logger := zap.Must(cfg.Build(options...))

		return logger
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Level = lvl
		logger := zap.Must(cfg.Build())

		return logger
	}

}

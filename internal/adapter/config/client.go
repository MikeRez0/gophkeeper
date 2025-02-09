package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

// ConfigClient - config params for client app.
//
// Config file example:
//
//	{
//	    "address": "localhost:8080", // аналог переменной окружения ADDRESS или флага -a
//	}
type ConfigClient struct {
	App          *App     `json:"app"`
	HostString   string   `env:"ADDRESS" json:"address"`
	LogLevel     string   `env:"LOG_LEVEL"`
	SyncInterval Duration `json:"sync_interval"` //env:"SYNC_INTERVAL"
	RateLimit    int      `env:"RATE_LIMIT"`
	GRPC         bool     `env:"GRPC_MODE" json:"grpc_mode"`
}

// NewConfigClient - Parse and create new client app config.
func NewConfigClient() (*ConfigClient, error) {
	// null config
	config := ConfigClient{
		App:          &App{LogLevel: "debug", Mode: AppModeDevelop},
		HostString:   `localhost:8080`,
		SyncInterval: Duration{2 * time.Second},
		RateLimit:    3,
		GRPC:         false,
	}

	err := loadConfigFile(&config)
	if err != nil {
		return nil, fmt.Errorf("error loading config file:%w", err)
	}

	var syncInterval int
	// cmd string params
	flag.String("c", "", cConfigFilenameUsage)
	flag.StringVar(&config.HostString, "a", config.HostString, "HTTP/gRPC server endpoint")
	flag.BoolVar(&config.GRPC, "g", config.GRPC, "Enable gRPC Mode")
	flag.IntVar(&syncInterval, "s", -1, "Sync interval")
	flag.IntVar(&config.RateLimit, "l", config.RateLimit, "Rate limit")
	flag.StringVar(&config.App.LogLevel, "log", config.App.LogLevel, "Log level")
	flag.Parse()

	if syncInterval != -1 {
		config.SyncInterval = Duration{time.Duration(syncInterval) * time.Second}
	}

	// environment override
	err = env.Parse(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing env config: %w", err)
	}

	err = lookupEnvDuration("SYNC_INTERVAL", &config.SyncInterval)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

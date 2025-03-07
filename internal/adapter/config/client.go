package config

import (
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
	ConfigFile          string
	App                 *App      `json:"app"`
	Database            *Database `json:"database"`
	HostString          string    `env:"ADDRESS" json:"address"`
	SyncInterval        time.Duration
	SyncIntervalSeconds int  `json:"sync_interval"` //env:"SYNC_INTERVAL"
	GRPC                bool `env:"GRPC_MODE" json:"grpc_mode"`
}

// NewConfigClient - Parse and create new client app config.
func NewConfigClient() *ConfigClient {
	// null config
	config := ConfigClient{
		App:                 &App{LogLevel: "debug", Mode: AppModeDevelop},
		HostString:          `localhost:8080`,
		SyncInterval:        2 * time.Second,
		SyncIntervalSeconds: 2,
		GRPC:                false,
		Database:            &Database{DSN: "keychain.db", Driver: "sqlite3"},
	}
	return &config
}

func (config *ConfigClient) LoadConfigFile() error {
	err := loadConfigFile(config)
	if err != nil {
		return fmt.Errorf("error loading config file:%w", err)
	}
	return nil
}
func (config *ConfigClient) Parse() error {
	// cmd string params
	// flag.String("c", "", cConfigFilenameUsage)
	// flag.StringVar(&config.HostString, "a", config.HostString, "HTTP/gRPC server endpoint")
	// flag.BoolVar(&config.GRPC, "g", config.GRPC, "Enable gRPC Mode")
	// flag.IntVar(&config.SyncIntervalSeconds, "s", config.SyncIntervalSeconds, "Sync interval")
	// flag.StringVar(&config.App.LogLevel, "log", config.App.LogLevel, "Log level")
	// flag.Parse()

	// environment override
	err := env.Parse(config)
	if err != nil {
		return fmt.Errorf("error parsing env config: %w", err)
	}

	config.SyncInterval = time.Second * time.Duration(config.SyncIntervalSeconds)

	return nil
}

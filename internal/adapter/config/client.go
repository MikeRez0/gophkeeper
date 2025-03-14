package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

// ConfigClient - config params for client app.
//
//	Param				Env			Flag		JSON			Default
//	---------------------------------------------------------------------------------------------------------------------
//	ConfigFile			-			c, config	-			-
//	App.*				-			-		app			-
//	App.LogLevel			LOG_LEVEL		l, log		log_level		debug
//	App.Mode			APP_MODE		-		mode			DEV
//	Database.*			-			-		database		-
//	Database.DSN			DATABASE_URI		-		database_dsn		-
//	Database.Driver			DATABASE_DRIVER		-		database_driver		-
//	HostString			ADDRESS			a, address	address			localhost:8080
//	TLSCertFile			TLS_CERTFILE		-		tls_cert		-
//	SyncInterval			-			-		-			2 * time.Second
//	SyncIntervalSeconds		-			s, sync		sync_interval		2
//
// Config file example:
//
//	{
//	  "database": {
//	    "database_dsn": "keychain.db",
//	    "database_driver": "sqlite3"
//	  },
//	  "address": "https://127.0.0.1:8888",
//	  "tls_cert": ".var/cert/cert.pem",
//	  "app": {
//	    "log_level": "info"
//	  }
//	}
//
// Params priority (ascending):
//  1. default value
//  2. config file value
//  3. flag from command line
//  4. environment variable.
//
//nolint:dupword // ok
type ConfigClient struct {
	ConfigFile          string
	App                 *App          `json:"app"`
	Database            *Database     `json:"database"`
	HostString          string        `env:"ADDRESS" json:"address"`
	TLSCertFile         string        `env:"TLS_CERTFILE" json:"tls_cert"`
	SyncInterval        time.Duration ``
	SyncIntervalSeconds int           `json:"sync_interval"`
	GRPC                bool          `env:"GRPC_MODE" json:"grpc_mode"`
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

// LoadConfigFile loads config params from file.
func (config *ConfigClient) LoadConfigFile() error {
	err := loadConfigFile(config)
	if err != nil {
		return fmt.Errorf("error loading config file:%w", err)
	}
	return nil
}

// Parse parses environment config params and runs final calculations.
func (config *ConfigClient) Parse() error {
	// environment override
	err := env.Parse(config)
	if err != nil {
		return fmt.Errorf("error parsing env config: %w", err)
	}

	config.SyncInterval = time.Second * time.Duration(config.SyncIntervalSeconds)

	return nil
}

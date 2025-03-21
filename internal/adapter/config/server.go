package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	_ "github.com/mattn/go-sqlite3"
)

// ConfigServer - config params for server aoo.
//
//	Param				Env			Flag		JSON			Default
//	---------------------------------------------------------------------------------------------------------------------
//	ConfigFile			-			c, config	-			-
//	App.*				-			-		app			-
//	App.LogLevel			LOG_LEVEL		l		log_level		debug
//	App.Mode			APP_MODE		-		mode			DEV
//	Database.*			-			-		database		-
//	Database.DSN			DATABASE_URI		d		database_dsn		-
//	Database.Driver			DATABASE_DRIVER		i		database_driver		-
//	HTTP.*				-			-		-			-
//	HTTP.HostString			RUN_ADDRESS		a		address			localhost:8080
//	HTTP.TLSCertFile		TLS_CERTFILE		-		tls_cert		-
//	HTTP.TLSKey			TLS_KEYFILE		-		tls_key			-
//
// Config file example:
//
//	{
//	  "database": {
//	    "database_dsn": "postgresql://gophkeeper:gophkeeper@localhost:5432/gophkeeper_db?sslmode=disable",
//	    "database_driver": "postgresql"
//	  },
//	  "http": {
//	    "address": "localhost:8080",
//	    "tls_key": ".var/cert/key.pem",
//	    "tls_cert": ".var/cert/cert.pem"
//	  },
//	  "app": {
//	    "log_level": "debug"
//	  }
//	}
//
// Params priority (ascending):
//  1. default value
//  2. config file value
//  3. flag from command line
//  4. environment variable.
type ConfigServer struct {
	Database *Database `json:"database"`
	HTTP     *HTTP     `json:"http"`
	App      *App      `json:"app"`
}

// HTTP section configuration.
type HTTP struct {
	HostString  string `env:"RUN_ADDRESS" json:"address"`   // server hostname
	TLSCertFile string `env:"TLS_CERTFILE" json:"tls_cert"` // TLS-connection cert file
	TLSKeyFile  string `env:"TLS_KEYFILE" json:"tls_key"`   // TLS-connection private key file
}

// NewConfigServer creates and parse new config for server.
func NewConfigServer() (*ConfigServer, error) {
	config := ConfigServer{
		Database: &Database{Driver: "sqlite3"},
		HTTP: &HTTP{
			HostString: `localhost:8080`,
		},
		App: &App{
			LogLevel: "debug",
			Mode:     AppModeDevelop,
		},
	}

	err := loadConfigFile(&config)
	if err != nil {
		return nil, fmt.Errorf("error loading config file:%w", err)
	}

	flag.String("c", "", cConfigFilenameUsage)
	flag.StringVar(&config.Database.DSN, "d", config.Database.DSN, "Database string")
	flag.StringVar(&config.Database.Driver, "i", config.Database.Driver, "Database driver")
	flag.StringVar(&config.HTTP.HostString, "a", config.HTTP.HostString, "HTTP server endpoint")
	flag.StringVar(&config.App.LogLevel, "l", config.App.LogLevel, "Log level")
	appModeStr := flag.String("m", string(config.App.Mode), "PROD / DEV")
	flag.Parse()

	config.App.Mode = AppMode(*appModeStr)

	err = env.Parse(config.Database)
	if err != nil {
		return nil, fmt.Errorf("error parsing env database config: %w", err)
	}
	err = env.Parse(config.HTTP)
	if err != nil {
		return nil, fmt.Errorf("error parsing http config: %w", err)
	}
	err = env.Parse(config.App)
	if err != nil {
		return nil, fmt.Errorf("error parsing app config: %w", err)
	}

	return &config, nil
}

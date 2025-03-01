package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	_ "github.com/mattn/go-sqlite3"
)

// ConfigServer - config params for server.
//
// Config file example:
//
//	{
//		"database": {
//		    "database_dsn": ""
//		},
//		"http":{
//		    "address": "localhost:8080"
//		}
//		"app":{
//		    "log_leve": "debug",
//			"mode": "DEV"
//		}
//	}
type ConfigServer struct {
	Database *Database `json:"database"`
	HTTP     *HTTP     `json:"http"`
	App      *App      `json:"app"`
}

type HTTP struct {
	HostString string `env:"RUN_ADDRESS" json:"address"`
}

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

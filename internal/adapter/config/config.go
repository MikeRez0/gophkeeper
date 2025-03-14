// Package config contains configuration params for client and server applications.
package config

// AppMode - run mode for application.
//
//   - PROD - production mode.
//   - DEV - developer mode.
type AppMode string

const (
	AppModeProduction AppMode = "PROD"
	AppModeDevelop    AppMode = "DEV"
)

// App - application params.
//
//   - LogLevel.
//   - Mode - application mode.
type App struct {
	LogLevel string  `env:"LOG_LEVEL" json:"log_level"`
	Mode     AppMode `env:"APP_MODE" json:"mode"`
}

// Database - database params.
//
//   - DSN - database connection string.
//   - Driver - database driver. "postgresql" and "sqlite3" supported.
type Database struct {
	DSN    string `env:"DATABASE_URI" json:"database_dsn"`
	Driver string `env:"DATABASE_DRIVER" json:"database_driver"`
}

package config

type AppMode string

const (
	AppModeProduction AppMode = "PROD"
	AppModeDevelop    AppMode = "DEV"
)

type App struct {
	LogLevel string  `env:"LOG_LEVEL" json:"log_level"`
	Mode     AppMode `env:"APP_MODE" json:"mode"`
}

type Database struct {
	DSN    string `env:"DATABASE_URI" json:"database_dsn"`
	Driver string `env:"DATABASE_DRIVER" json:"database_driver"`
}

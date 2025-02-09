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

package logger

import (
	"compass/config"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Configuration keys
var (
	LogFormatKey = config.ConfigKey("log.format")
)

func init() {
	// Set logger level field to severity for stack driver support
	zerolog.LevelFieldName = "severity"
	// Set defaults
	viper.SetDefault(LogFormatKey.String(), "json")
	// Bind environment variables
	viper.BindEnv(LogFormatKey.String())
}

// ListenAddress returns the configured gRPC server listen address
func LogFormat() string {
	return viper.GetString(LogFormatKey.String())
}

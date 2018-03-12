package logger

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Configuration keys
var (
	LogFormatKey = "log.format"
)

func init() {
	// Set logger level field to severity for stack driver support
	zerolog.LevelFieldName = "severity"
	// Set defaults
	viper.SetDefault(LogFormatKey, "json")
	// Bind environment variables
	viper.BindEnv(LogFormatKey)
}

// ListenAddress returns the configured gRPC server listen address
func LogFormat() string {
	return viper.GetString(LogFormatKey)
}

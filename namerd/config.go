package namerd

import (
	"compass/config"

	"github.com/spf13/viper"
)

// Configuration keys
var (
	HostKey   = config.ConfigKey("namerd.host")
	SchemeKey = config.ConfigKey("namerd.scheme")
)

func init() {
	// Set defaults
	viper.SetDefault(HostKey.String(), "127.0.0.1:4180")
	viper.SetDefault(SchemeKey.String(), "http")
	// Bind environment variables
	viper.BindEnv(HostKey.String(), SchemeKey.String())
}

// Host returns the configured namerd host
func Host() string {
	return viper.GetString(HostKey.String())
}

// Scheme returns the configured namerd host
func Scheme() string {
	return viper.GetString(SchemeKey.String())
}

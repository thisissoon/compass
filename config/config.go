package config

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// application name
var configName = "compass"

// configuration path viper lookup keys
const (
	CONFIG_PATH_KEY = "config"
)

// init sets default configuration file settings such as path look
// up values and environment variable binding
func init() {
	// Config file lookup locations
	viper.SetConfigType("toml")
	viper.SetConfigName("compass")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("$HOME/.config/compass")
	viper.AddConfigPath("/etc/compass")
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("compass")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Environment variable binding
	viper.BindEnv(CONFIG_PATH_KEY)
}

// FromFile reads configuration from a file, bind a CLI flag to
func FromFile() error {
	path := viper.GetString(CONFIG_PATH_KEY)
	if path != "" {
		viper.SetConfigFile(path)
	}
	return viper.ReadInConfig()
}

// BindPFlag binds a config key to a CLI pflag
func BindPFlags(flags map[string]*pflag.Flag) {
	for k, f := range flags {
		viper.BindPFlag(k, f)
	}
}

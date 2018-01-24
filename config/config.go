package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// compile time variables
var (
	// default config file name set at compile time
	filename = "compass"
	// default environment var prefix set at compile time
	envprefix = "compass"
)

// Set this to override the default config file lookup to load
// a specific config file
var Path string

// init sets default configuration file settings such as path look
// up values and environment variable binding
func init() {
	// Config file lookup locations
	viper.SetConfigType("toml")
	viper.SetConfigName(filename)
	viper.AddConfigPath(filepath.Join("$HOME", ".config"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config", "compass"))
	viper.AddConfigPath(filepath.Join("etc", "compass"))
	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envprefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// A ConfigKey is used to lookup configuration values
type ConfigKey string

// String returns the string value of the config key
func (k ConfigKey) String() string {
	return string(k)
}

// FromFile reads configuration from a file, bind a CLI flag to
func Read() error {
	if Path != "" {
		viper.SetConfigFile(Path)
	}
	return viper.ReadInConfig()
}

// BindFlag binds a CLI flag to the provided config key
func BindFlag(k ConfigKey, f *pflag.Flag) {
	viper.BindPFlag(k.String(), f)
}

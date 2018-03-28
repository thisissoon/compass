package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Configuration keys
var (
	// configuration file path
	configPathKey = "config.path"
	// Logging configuration
	logFormatKey = "log.format"
	// needle configuration
	needleNamespaceKey = "needle.namespace"
)

// init sets default configuration file settings such as path look
// up values and environment variable binding
// /etc/compass/compass.toml
// ~/.config/compass.toml
// ~/.config/compass/compass.toml
// ~/.compass/compass.toml
// $(pwd)/.compass/compass.toml
func init() {
	// Default logging configuration
	viper.SetDefault(logFormatKey, "json")
	// Kubernetets configuration defaults
	viper.SetDefault(needleNamespaceKey, "kube-system")
	// Config file lookup locations
	viper.SetConfigType("toml")
	viper.SetConfigName("compass")
	viper.AddConfigPath(filepath.Join("etc", "compass"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config", "compass"))
	viper.AddConfigPath(filepath.Join("$HOME", ".compass"))
	viper.AddConfigPath(filepath.Join("$PWD", ".compass"))
	// Environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("compass")
	viper.AutomaticEnv()
	viper.BindEnv(
		logFormatKey,
		needleNamespaceKey)
}

// readConfig loads configuraion from a file
func readConfig() error {
	if viper.GetString(configPathKey) != "" {
		viper.SetConfigFile(viper.GetString(configPathKey))
	}
	return viper.ReadInConfig()
}

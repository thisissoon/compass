package main

import (
	"compass/logger"
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
	// k8s configuration
	kubeContextKey   = "kube.context"
	kubeNamespaceKey = "kube.namespasce"
)

// init sets default configuration file settings such as path look
// up values and environment variable binding
// /etc/compass/compass.toml
// ~/.config/compass.toml
// ~/.config/compass/compass.toml
// ~/.compass/compass.toml
func init() {
	// Default logging configuration
	viper.SetDefault(logFormatKey, "json")
	// Kubernetets configuration defaults
	viper.SetDefault(kubeContextKey, "")
	viper.SetDefault(kubeNamespaceKey, "kube-system")
	// Config file lookup locations
	viper.SetConfigType("toml")
	viper.SetConfigName("compass")
	viper.AddConfigPath(filepath.Join("etc", "compass"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config", "compass"))
	viper.AddConfigPath(filepath.Join("$HOME", ".compass"))
	// Environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("compass")
	viper.AutomaticEnv()
}

// readConfig loads configuraion from a file
func readConfig() {
	logger := logger.New()
	if viper.GetString(configPathKey) != "" {
		viper.SetConfigFile(viper.GetString(configPathKey))
	}
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn().Err(err).Msg("error reading configuration file")
	}
}

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
	// grpc configuration
	grpcListenKey = "grpc.listen"
	// logging configuration
	logFormatKey = "log.format"
	// db configuration
	dbHostKey    = "psql.host"
	dbNameKey    = "psql.db"
	dbUserKey    = "psql.username"
	dbPassKey    = "psql.password"
	dbSSLModeKey = "psql.sslmode"
	// namerd configuration
	namerdHostKey   = "namerd.host"
	namerdSchemeKey = "namerd.scheme"
)

// init sets default configuration file settings such as path look
// up values and environment variable binding
// /etc/needle/needle.toml
// ~/.config/needle.toml
// ~/.config/needle/needle.toml
// ~/.needle/needle.toml
func init() {
	// Default gRPC configuration
	viper.SetDefault(grpcListenKey, ":5000")
	// Default logging configuration
	viper.SetDefault(logFormatKey, "json")
	// Default psql configuration
	viper.SetDefault(dbHostKey, "localhost:5432")
	viper.SetDefault(dbNameKey, "needle")
	viper.SetDefault(dbUserKey, "postgres")
	viper.SetDefault(dbPassKey, "postgres")
	viper.SetDefault(dbSSLModeKey, "disable")
	// Namerd default configuration'
	viper.SetDefault(namerdHostKey, "127.0.0.1:4180")
	viper.SetDefault(namerdSchemeKey, "http")
	// Config file lookup locations
	viper.SetConfigType("toml")
	viper.SetConfigName("needle")
	viper.AddConfigPath(filepath.Join("etc", "needle"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config"))
	viper.AddConfigPath(filepath.Join("$HOME", ".config", "needle"))
	viper.AddConfigPath(filepath.Join("$HOME", ".needle"))
	viper.AddConfigPath(filepath.Join("$PWD", ".needle"))
	// Environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("needle")
	viper.AutomaticEnv()
}

// readConfig loads configuraion from a file
func readConfig() error {
	if viper.GetString(configPathKey) != "" {
		viper.SetConfigFile(viper.GetString(configPathKey))
	}
	return viper.ReadInConfig()
}

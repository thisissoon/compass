package psql

import (
	"compass/config"
	"fmt"

	"github.com/spf13/viper"
)

// Configuration keys
var (
	hostKey     = config.ConfigKey("psql.host")
	dbKey       = config.ConfigKey("psql.db")
	usernameKey = config.ConfigKey("psql.username")
	passwordKey = config.ConfigKey("psql.password")
)

func init() {
	// Set defaults
	viper.SetDefault(hostKey.String(), "locahost:5432")
	viper.SetDefault(dbKey.String(), "needle")
	viper.SetDefault(usernameKey.String(), "postgres")
	viper.SetDefault(passwordKey.String(), "postgres")
	// Bind environment variables
	viper.BindEnv(hostKey.String())
	viper.BindEnv(dbKey.String())
	viper.BindEnv(usernameKey.String())
	viper.BindEnv(passwordKey.String())
}

func host() string {
	return viper.GetString(hostKey.String())
}

func db() string {
	return viper.GetString(dbKey.String())
}

func username() string {
	return viper.GetString(usernameKey.String())
}

func password() string {
	return viper.GetString(passwordKey.String())
}

// DSN returns the configured postgres dsn address
func DSN() string {
	dsn := "postgres://%s:%s@%s/%s?sslmode=disable"
	return fmt.Sprintf(dsn, username(), password(), host(), db())
}

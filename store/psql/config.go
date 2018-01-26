package psql

import (
	"compass/config"

	"github.com/spf13/viper"
)

// Default Data Source Name
const defaultDsn = "postgres://postgres:postgres@localhost:5432/needle?sslmode=disable"

// Configuration keys
var (
	DSNKey = config.ConfigKey("psql.dsn")
)

func init() {
	// Set defaults
	viper.SetDefault(DSNKey.String(), defaultDsn)
	// Bind environment variables
	viper.BindEnv(DSNKey.String())
}

// DSN returns the configured postgres dsn address
func DSN() string {
	return viper.GetString(DSNKey.String())
}

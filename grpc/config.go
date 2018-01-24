package grpc

import (
	"compass/config"

	"github.com/spf13/viper"
)

// Configuration keys
var (
	ListenAddressConfigKey = config.ConfigKey("grpc.listen_address")
)

func init() {
	// Set defaults
	viper.SetDefault(ListenAddressConfigKey.String(), ":5000")
	// Bind environment variables
	viper.BindEnv(ListenAddressConfigKey.String())
}

// ListenAddress returns the configured gRPC server listen address
func ListenAddress() string {
	return viper.GetString(ListenAddressConfigKey.String())
}

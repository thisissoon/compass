package grpc

import (
	"compass/config"

	"github.com/spf13/viper"
)

// Configuration keys
var (
	ListenAddressConfigKey = config.ConfigKey("grpc.listen_address")
	ClientAddressKey       = config.ConfigKey("grpc.client_address")
)

func init() {
	// Set defaults
	viper.SetDefault(ListenAddressConfigKey.String(), ":5000")
	viper.SetDefault(ClientAddressKey.String(), "localhosts:5000")
	// Bind environment variables
	viper.BindEnv(ListenAddressConfigKey.String())
	viper.BindEnv(ClientAddressKey.String())
}

// ListenAddress returns the configured gRPC server listen address
func ListenAddress() string {
	return viper.GetString(ListenAddressConfigKey.String())
}

// ClientAddress returns the configured gRPC client address
func ClientAddress() string {
	return viper.GetString(ClientAddressKey.String())
}

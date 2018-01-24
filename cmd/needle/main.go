package main

import (
	"os"
	"os/signal"
	"syscall"

	"compass/config"
	"compass/grpc"
	"compass/logger"

	"github.com/spf13/cobra"
)

var sigC = make(chan os.Signal)

// Application entry point
func main() {
	needleCmd().Execute()
}

// New constructs a new CLI interface for execution
func needleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "needle",
		Short: "Start Needle, the gRPC server for the Compass client.",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(startNeedle())
		},
	}
	// Global flags
	pflags := cmd.PersistentFlags()
	pflags.String("log-format", "", "log format [console|json]")
	pflags.StringVarP(&config.Path, "config", "c", "", "Path to configuration file")
	// Bind persistent flags
	config.BindFlag(logger.LogFormatKey, pflags.Lookup("log-format"))
	// Local Flags
	flags := cmd.Flags()
	flags.StringP("listen", "l", "", "server listen address")
	// Bind local flags to config options
	config.BindFlag(grpc.ListenAddressConfigKey, flags.Lookup("listen"))
	// Add sub commands
	cmd.AddCommand(versionCmd())
	return cmd
}

// startNeedle starts the needle gRPC server
// returns os exit code
func startNeedle() int {
	config.Read()
	logger := logger.New()
	srv := grpc.NewServer(grpc.WithAddress(grpc.ListenAddress()))
	addr, errC := srv.Serve()
	logger.Debug().Str("address", addr.String()).Msg("gRPC server started")
	defer srv.Stop()
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case sig := <-sigC:
		logger.Debug().Str("signal", sig.String()).Msg("recieved OS signal")
		return 0
	case err := <-errC:
		logger.Debug().Err(err).Msg("runtime error")
		return 1
	}
}

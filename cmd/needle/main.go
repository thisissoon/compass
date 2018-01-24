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
		Short: "Needle is the gRPC server for the compass client.",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(startNeedle())
		},
	}
	// Global flags
	pflags := cmd.PersistentFlags()
	pflags.StringP("config-file", "c", "", "path to configuration file")
	pflags.String("log-format", "", "log format [console|json]")
	// Local Flags
	flags := cmd.Flags()
	flags.StringP("listen", "l", "", "server listen address")
	// Bind flags to config options
	config.BindFlag(grpc.ListenAddressConfigKey, flags.Lookup("listen"))
	// Add sub commands
	cmd.AddCommand(versionCmd())
	return cmd
}

// startNeedle starts the needle gRPC server
// returns os exit code
func startNeedle() int {
	config.FromFile()
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

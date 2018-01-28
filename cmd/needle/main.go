package main

import (
	"os"
	"os/signal"
	"syscall"

	"compass/config"
	"compass/grpc"
	"compass/logger"
	"compass/namerd"
	"compass/needle"
	"compass/store/psql"
	"compass/sync"

	"github.com/jmoiron/sqlx"
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
	cmd.AddCommand(
		migrateCmd(),
		versionCmd())
	return cmd
}

// loadConfig loads configuraion
func loadConfig() {
	logger := logger.New()
	if err := config.Read(); err != nil {
		logger.Error().Err(err).Msg("error loading configuration")
	}
}

// startNeedle starts the needle gRPC server
// returns os exit code
func startNeedle() int {
	loadConfig()
	log := logger.New()
	db, err := sqlx.Open("postgres", psql.DSN())
	if err != nil {
		log.Error().Err(err).Msg("failed to open database connection")
		return 1
	}
	defer db.Close()

	syncer := sync.New(namerd.New())
	go syncer.Start()
	defer syncer.Stop()

	srv := grpc.NewServer(needle.NewService(psql.New(db)))
	addr, errC := srv.Serve()
	log.Debug().Str("address", addr.String()).Msg("gRPC server started")
	defer srv.Stop()
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case sig := <-sigC:
		log.Debug().Str("signal", sig.String()).Msg("recieved OS signal")
		return 0
	case err := <-errC:
		log.Debug().Err(err).Msg("runtime error")
		return 1
	}
}

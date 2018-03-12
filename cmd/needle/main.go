package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"compass/grpc"
	"compass/k8s"
	"compass/logger"
	"compass/namerd"
	"compass/needle"
	"compass/store/psql"
	"compass/sync"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// OS Signal Channel
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
	pflags.String("config-path", "", "Path to configuration file")
	pflags.String("log-format", "", "Log output format [console|json]")
	// Bind global flags to viper configs
	viper.BindPFlag(configPathKey, pflags.Lookup("config-path"))
	viper.BindPFlag(logFormatKey, pflags.Lookup("log-format"))
	// Local Flags
	flags := cmd.Flags()
	flags.StringP("listen", "l", "", "gRPC server listen address, e.g :5000")
	// Bind local flags to config options
	viper.BindPFlag(grpcListenKey, pflags.Lookup("listen"))
	// Add sub commands
	cmd.AddCommand(
		migrateCmd(),
		versionCmd())
	return cmd
}

// dbDSN returns the database connection url
func dbDSN() psql.DSN {
	return psql.DSN{
		Name:     viper.GetString(dbNameKey),
		Host:     viper.GetString(dbHostKey),
		Username: viper.GetString(dbUserKey),
		Password: viper.GetString(dbPassKey),
		SSLMode:  viper.GetString(dbSSLModeKey),
	}
}

// openDB opens a database connection
func openDB() (*sqlx.DB, error) {
	return psql.Open(dbDSN())
}

// startNeedle starts the needle gRPC server
// returns os exit code
func startNeedle() int {
	readConfig()
	ctx, cancel := context.WithCancel(context.Background())
	log := logger.New()
	db, err := openDB()

	if err != nil {
		log.Error().Err(err).Msg("failed to open database connection")
		return 1
	}
	defer db.Close()
	store := psql.New(db)
	nd := namerd.New(
		namerd.WithHost(viper.GetString(namerdHostKey)),
		namerd.WithScheme(viper.GetString(namerdSchemeKey)))
	syncer := sync.New(nd, store, store)
	go syncer.Start(ctx)
	kcc, err := k8s.Clientset()
	if err != nil {
		log.Error().Err(err).Msg("failed to obtain kubernetes configuration")
		return 1
	}
	kc := k8s.New(kcc)
	srv := grpc.NewServer(
		needle.NewService(store, kc),
		grpc.WithAddress(viper.GetString(grpcListenKey)))
	errC := srv.Serve()
	defer srv.Stop()
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	select {
	case sig := <-sigC:
		log.Debug().Str("signal", sig.String()).Msg("recieved OS signal")
		return 0
	case err := <-errC:
		log.Debug().Err(err).Msg("runtime error")
		return 1
	}
}

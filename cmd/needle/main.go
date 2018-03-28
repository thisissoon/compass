package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"compass/pkg/kube"
	"compass/pkg/namerd"
	"compass/pkg/server"
	"compass/pkg/service"
	"compass/pkg/store/psql"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// OS Signal Channel
var sigC = make(chan os.Signal)

// Common Logger
var log = zerolog.New(os.Stdout)

// Application entry point
func main() {
	needleCmd().Execute()
}

// setup reads configuration and updates logger output
func setup() {
	readConfig()
	switch viper.GetString(logFormatKey) {
	case "discard":
		log = log.Output(ioutil.Discard)
	case "console":
		log = log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})
	}
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

// New constructs a new CLI interface for execution
func needleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "needle",
		Short: "Start Needle, the gRPC server for the Compass client.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return startNeedle()
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
	viper.BindPFlag(grpcListenKey, flags.Lookup("listen"))
	// Add sub commands
	cmd.AddCommand(
		migrateCmd(),
		versionCmd())
	return cmd
}

// startNeedle starts the needle gRPC server
// returns os exit code
func startNeedle() error {
	setup()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.WithContext(ctx)
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	store := psql.New(db)
	nd := namerd.New(
		namerd.WithHost(viper.GetString(namerdHostKey)),
		namerd.WithScheme(viper.GetString(namerdSchemeKey)))
	syncer := namerd.Syncer(nd, store, store)
	go syncer.Start(ctx)
	kcc, err := kube.Clientset()
	if err != nil {
		return err
	}
	srv := server.New(
		service.New(store, kcc),
		server.WithAddress(viper.GetString(grpcListenKey)))
	errC := srv.Serve(log)
	defer srv.Stop()
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	select {
	case <-sigC:
		return nil
	case err := <-errC:
		return err
	}
}

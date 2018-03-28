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
	"github.com/mattes/migrate"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// OS Signal Channel
var (
	sigCh = make(chan os.Signal)
	errCh = make(chan error)
)

// Common Logger
var log = zerolog.New(os.Stdout)

// DB Connection
var (
	db       *sqlx.DB
	migrator *migrate.Migrate
)

// Application entry point
func main() {
	needleCmd().Execute()
}

// setup reads configuration and updates logger output
func setup() error {
	var err error
	readConfig()
	switch viper.GetString(logFormatKey) {
	case "discard":
		log = log.Output(ioutil.Discard)
	case "console":
		log = log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})
	}
	db, err = psql.Open(psql.DSN{
		Name:     viper.GetString(dbNameKey),
		Host:     viper.GetString(dbHostKey),
		Username: viper.GetString(dbUserKey),
		Password: viper.GetString(dbPassKey),
		SSLMode:  viper.GetString(dbSSLModeKey),
	})
	if err != nil {
		return err
	}
	migrator, err = psql.NewMigratorFromDB(
		db.DB,
		psql.MigrateWithLog(&psql.MigrateLogger{zerolog.Nop()}))
	if err != nil {
		return err
	}
	return nil
}

func teardown() error {
	if db != nil {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

// New constructs a new CLI interface for execution
func needleCmd() *cobra.Command {
	var listen string
	cmd := &cobra.Command{
		Use:   "needle",
		Short: "Start Needle, the gRPC server for the Compass client.",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setup()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return start(listen)
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			return teardown()
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
	flags.StringVarP(&listen, "listen", "l", "", "gRPC server listen address, e.g :5000")
	// Add sub commands
	cmd.AddCommand(
		migrateCmd(),
		versionCmd())
	return cmd
}

// startNeedle starts the needle gRPC server
// returns os exit code
func start(address string) error {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.WithContext(ctx)
	store := psql.New(db)
	nd := namerd.New(
		namerd.WithHost(viper.GetString(namerdHostKey)),
		namerd.WithScheme(viper.GetString(namerdSchemeKey)))
	syncer := namerd.Syncer(nd, store, store)
	go syncer.Start(ctx)
	cs, err := kube.Clientset()
	if err != nil {
		return err
	}
	srv := server.New(server.WithLogger(log), server.WithAddress(address))
	go func() {
		errCh <- srv.Serve(
			service.NewDentryService(store, cs),
			service.NewDataMigrationService(migrator))
	}()
	defer srv.Stop()
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	select {
	case <-sigCh:
		return nil
	case err := <-errCh:
		return err
	}
}

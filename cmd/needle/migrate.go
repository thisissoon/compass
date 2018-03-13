package main

import (
	"fmt"
	"os"

	"compass/store/psql"

	"github.com/spf13/cobra"
)

// migrateCmd
func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
		Run: func(*cobra.Command, []string) {
		},
	}
	cmd.AddCommand(
		migrateUpCmd(),
		migrateDownCmd(),
		migrateVersionCmd(),
		migrateDropCmd())
	return cmd
}

// migrateUpCmd
func migrateUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run database upgrade migrations.",
		Run: func(*cobra.Command, []string) {
			os.Exit(migrateUp())
		},
	}
	return cmd
}

// migrateUp runs database upgrade migrations
func migrateUp() int {
	setup()
	dsn := dbDSN()
	log.Debug().Str("dsn", dsn.String()).Msg("opening database connection")
	m, err := psql.NewMigrator(dsn, &psql.MigrateLogger{log})
	if err != nil {
		log.Error().Err(err).Msg("failed to create database migrator")
		return 1
	}
	defer m.Close()
	if err := m.Up(); err != nil {
		log.Error().Err(err).Msg("failed to run database upgrade")
		return 1
	}
	return 0
}

// migrateDownCmd
func migrateDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Run downgrade migrations.",
		Run: func(*cobra.Command, []string) {
			os.Exit(migrateDown())
		},
	}
	return cmd
}

// migrateDown runs database downgrade migrations
func migrateDown() int {
	setup()
	dsn := dbDSN()
	log.Debug().Str("dsn", dsn.String()).Msg("opening database connection")
	m, err := psql.NewMigrator(dsn, &psql.MigrateLogger{log})
	if err != nil {
		log.Error().Err(err).Msg("failed to create database migrator")
		return 1
	}
	defer m.Close()
	if err := m.Down(); err != nil {
		log.Error().Err(err).Msg("failed to run database downgrade")
		return 1
	}
	return 0
}

// migrateVersionCmd
func migrateVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display current migration version.",
		Run: func(*cobra.Command, []string) {
			os.Exit(migrateVersion())
		},
	}
	return cmd
}

// migrateVersion prints the current database version
func migrateVersion() int {
	setup()
	dsn := dbDSN()
	log.Debug().Msg("hello")
	log.Debug().Str("dsn", dsn.String()).Msg("opening database connection")
	m, err := psql.NewMigrator(dsn, &psql.MigrateLogger{log})
	if err != nil {
		log.Error().Err(err).Msg("failed to create database migrator")
		return 1
	}
	defer m.Close()
	version, dirty, err := m.Version()
	if err != nil {
		log.Error().Err(err).Msg("failed to get database migration version")
	}
	fmt.Println("Version:", version)
	fmt.Println("Dirty:", dirty)
	return 0
}

// migrateDropCmd
func migrateDropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drop database migrations.",
		Run: func(*cobra.Command, []string) {
			os.Exit(migrateDrop())
		},
	}
	return cmd
}

// migrateDrop prints the current database version
func migrateDrop() int {
	readConfig()
	dsn := dbDSN()
	log.Debug().Str("dsn", dsn.String()).Msg("opening database connection")
	m, err := psql.NewMigrator(dsn, &psql.MigrateLogger{log})
	if err != nil {
		log.Error().Err(err).Msg("failed to create database migrator")
		return 1
	}
	defer m.Close()
	if err := m.Drop(); err != nil {
		log.Error().Err(err).Msg("failed to get drop database migrations")
		return 1
	}
	return 0
}

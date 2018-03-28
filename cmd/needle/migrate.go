package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// migrateCmd
func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage database migrations",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return setup()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return teardown()
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
		RunE: func(*cobra.Command, []string) error {
			return migrator.Up()
		},
	}
	return cmd
}

// migrateDownCmd
func migrateDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Run downgrade migrations.",
		RunE: func(*cobra.Command, []string) error {
			return migrator.Down()
		},
	}
	return cmd
}

// migrateVersionCmd
func migrateVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display current migration version.",
		RunE: func(*cobra.Command, []string) error {
			v, d, err := migrator.Version()
			if err != nil {
				return err
			}
			fmt.Println("Version:", v)
			fmt.Println("Dirty:", d)
			return nil
		},
	}
	return cmd
}

// migrateDropCmd
func migrateDropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drop database migrations.",
		RunE: func(*cobra.Command, []string) error {
			return migrator.Drop()
		},
	}
	return cmd
}

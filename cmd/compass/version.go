package main

import (
	"compass/pkg/version"

	"github.com/spf13/cobra"
)

// versionCmd returns a CLI command that when run prints
// the application build version, commit and time
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints version information",
		Run: func(*cobra.Command, []string) {
			version.Print()
		},
	}
}

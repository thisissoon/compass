package main

import (
	"fmt"
	"time"

	"compass/pkg/version"

	"github.com/spf13/cobra"
)

// versionCmd returns a CLI command that when run prints
// the application build version, commit and time
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the build version",
		Run: func(*cobra.Command, []string) {
			fmt.Println("Version:", version.Version())
			fmt.Println("Commit:", version.Commit())
			fmt.Println("Built:", version.BuildTime().Format(time.RFC1123))
		},
	}
}

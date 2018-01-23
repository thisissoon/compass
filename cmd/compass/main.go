package main

import (
	"compass/config"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Application entry point
func main() {
	compassCmd().Execute()
}

// New constructs a new CLI interface for execution
func compassCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compass",
		Short: "Compass is a release management tool that handles managing namerd finagle delegation tables.",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
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
	config.BindPFlags(map[string]*pflag.Flag{
		config.CONFIG_PATH_KEY: pflags.Lookup("config-file"),
		config.LOG_FORMAT_KEY:  pflags.Lookup("log-format"),
	})
	// Add sub commands
	cmd.AddCommand(versionCmd())
	return cmd
}

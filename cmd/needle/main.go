package main

import (
	"compass/config"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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

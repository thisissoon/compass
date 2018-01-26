package main

import (
	"context"
	"fmt"
	"os"

	"compass/config"
	"compass/grpc"
	"compass/logger"
	needle "compass/proto/needle/v1"

	"github.com/spf13/cobra"
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
	pflags.String("log-format", "", "log format [console|json]")
	pflags.StringVarP(&config.Path, "config", "c", "", "Path to configuration file")
	// Bind persistent flags
	config.BindFlag(logger.LogFormatKey, pflags.Lookup("log-format"))
	// Add sub commands
	cmd.AddCommand(
		manageCmd(),
		versionCmd())
	return cmd
}

func manageCmd() *cobra.Command {
	var logicalName string
	var namespace string
	var dtab string
	var description string
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "Add / Update services compass will manage.",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(manage(
				logicalName,
				namespace,
				dtab,
				description,
			))
		},
	}
	cmd.Flags().StringVarP(&logicalName, "logical-name", "l", "", "Service logical name.")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes Namespace the service runs in.")
	cmd.Flags().StringVarP(&dtab, "dtab", "d", "", "Named delegation table name to place the dentry.")
	cmd.Flags().StringVarP(&description, "description", "D", "", "Optional service description.")
	return cmd
}

func manage(ln, ns, dt, dsc string) int {
	config.Read()
	log := logger.New()
	client, ok := grpc.NewClient(grpc.ClientAddress())
	if !ok {
		return 0
	}
	_, err := client.PutService(
		context.Background(),
		&needle.PutServiceRequest{
			Service: &needle.Service{
				LogicalName: ln,
				Dtab:        dt,
				Namespace:   ns,
				Description: dsc,
			},
		})
	if err != nil {
		log.Error().Err(err).Msg("failed to put service")
		fmt.Println("failed to create / update service")
		return 1
	}
	return 0
}

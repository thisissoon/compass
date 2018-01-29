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
		dentryCmd(),
		routeCmd(),
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
			os.Exit(putService(
				logicalName,
				namespace,
				dtab,
				description,
			))
		},
	}
	cmd.Flags().StringVarP(&logicalName, "logical-name", "l", "", "Service logical name.")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes Namespace the service runs in.")
	cmd.Flags().StringVarP(&description, "description", "D", "", "Optional service description.")
	return cmd
}

func putService(ln, ns, dt, dsc string) int {
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

func dentryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dentry",
		Short: "Add / Update manual dtab dentries.",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(dentryPutCmd())
	return cmd
}

func dentryPutCmd() *cobra.Command {
	var dtab string
	var prefix string
	var destination string
	var priority int32
	cmd := &cobra.Command{
		Use:   "put",
		Short: "Add / Update manual dtab dentries.",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(putDentry(
				dtab,
				prefix,
				destination,
				priority,
			))
		},
	}
	cmd.Flags().StringVarP(&dtab, "dtab", "d", "", "Delegation table the dentry should be in.")
	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "Rule prefix. e.g /svc/foo.")
	cmd.Flags().StringVarP(&destination, "destination", "D", "", "Rule destination. e.g /#/io.l5d.k8s/paralympics/http/foo.")
	cmd.Flags().Int32VarP(&priority, "priority", "P", 0, "Rule priority, higher = more important, e.g 100")
	return cmd
}

func putDentry(dt, p, dst string, pr int32) int {
	config.Read()
	log := logger.New()
	client, ok := grpc.NewClient(grpc.ClientAddress())
	if !ok {
		return 0
	}
	_, err := client.PutDentry(
		context.Background(),
		&needle.PutDentryRequest{
			Dentry: &needle.Dentry{
				Dtab:        dt,
				Prefix:      p,
				Destination: dst,
				Priority:    pr,
			},
		})
	if err != nil {
		log.Error().Err(err).Msg("failed to put dentry")
		fmt.Println("failed to create / update dentry")
		return 1
	}
	return 0
}

func routeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Update a services dentry.",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(routeVersionCmd())
	return cmd
}

func routeVersionCmd() *cobra.Command {
	var logicalName string
	var version string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Route a service a specifcly deployed version",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(routeVersion(logicalName, version))
		},
	}
	cmd.Flags().StringVarP(&logicalName, "logical-name", "l", "", "Service logical name.")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version e.g: develop, 1.2.3.")
	return cmd
}

func routeVersion(logicalName, version string) int {
	config.Read()
	log := logger.New()
	client, ok := grpc.NewClient(grpc.ClientAddress())
	if !ok {
		return 0
	}
	_, err := client.RouteToVersion(
		context.Background(),
		&needle.RouteToVersionRequest{
			LogicalName: logicalName,
			Version:     version,
		})
	if err != nil {
		log.Error().Err(err).Msg("failed to put dentry")
		fmt.Println("failed to create / update dentry")
		return 1
	}
	return 0
}

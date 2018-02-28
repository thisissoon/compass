package main

import (
	"fmt"
	"os"

	"compass/config"
	"compass/logger"
	"compass/needle/client"

	"github.com/spf13/cobra"
)

var options client.Options

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

	pflags.StringVar(&options.Context, "context", "", "Kubernetes context to use")
	pflags.StringVar(&options.Namespace, "namespace", "", "Kubernetes namespace needle runs in")

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
	var description string
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "Add / Update services compass will manage.",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(putService(
				logicalName,
				namespace,
				description,
			))
		},
	}
	cmd.Flags().StringVarP(&logicalName, "logical-name", "l", "", "Service logical name.")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes Namespace the service runs in.")
	cmd.Flags().StringVarP(&description, "description", "D", "", "Optional service description.")
	return cmd
}

func putService(ln, ns, dsc string) int {
	client, err := client.New(options)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	svc, err := client.PutService(ln, ns, dsc)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(fmt.Sprintf("Id: %s", svc.GetId()))
	fmt.Println(fmt.Sprintf("Logical Name: %s", svc.GetLogicalName()))
	fmt.Println(fmt.Sprintf("Kubernetes Namespace: %s", svc.GetNamespace()))
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
	cmd.AddCommand(dentryPutCmd(), deleteDentryCmd())
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
	client, err := client.New(options)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	dentry, err := client.PutDentry(dt, p, dst, pr)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(fmt.Sprintf("Id: %s", dentry.GetId()))
	fmt.Println(fmt.Sprintf("Delegation Table: %s", dentry.GetDtab()))
	fmt.Println(fmt.Sprintf("Priority: %d", dentry.GetPriority()))
	fmt.Println(fmt.Sprintf("Dentry: %s => %s", dentry.GetPrefix(), dentry.GetDestination()))
	return 0
}

func deleteDentryCmd() *cobra.Command {
	var (
		id     string
		dtab   string
		prefix string
	)
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete dentry by Id or Dtab & Prefix",
		Run: func(cmd *cobra.Command, _ []string) {
			if id != "" {
				os.Exit(deleteDentryById(id))
			} else if dtab != "" && prefix != "" {
				os.Exit(deleteDentryByPrefix(dtab, prefix))
			} else {
				os.Exit(deleteDentry())
			}
			cmd.Help()
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "Dentry ID in UUIDv4 format.")
	cmd.Flags().StringVarP(&dtab, "dtab", "d", "", "Delegation table the dentry is in, must also provide --prefix/-p.")
	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "Dentry prefix, must also provide --dtab/-d")
	return cmd
}

func deleteDentry() int {
	client, err := client.New(options)
	if err != nil {
		return 1
	}
	dtabs, err := client.DelegationTables()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(dtabs)
	return 0
}

func deleteDentryById(id string) int {
	client, err := client.New(options)
	if err != nil {
		return 1
	}
	ok, err := client.DeleteDentryById(id)
	if err != nil {
		return 1
	}
	if !ok {
		fmt.Println("Dentry was not deleted")
		return 1
	}
	fmt.Println("Dentry has been deleted")
	return 0
}

func deleteDentryByPrefix(dtab, prefix string) int {
	client, err := client.New(options)
	if err != nil {
		return 1
	}
	ok, err := client.DeleteDentryByPrefix(dtab, prefix)
	if err != nil {
		return 1
	}
	if !ok {
		fmt.Println("Dentry was not deleted")
		return 1
	}
	fmt.Println("Dentry has been deleted")
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
	client, err := client.New(options)
	if err != nil {
		return 1
	}
	if err := client.RouteToVersion(logicalName, version); err != nil {
		fmt.Println("Could not route to version")
		return 1
	}
	fmt.Println(fmt.Sprintf("Updated delegation table to route %s to %s", logicalName, version))
	return 0
}

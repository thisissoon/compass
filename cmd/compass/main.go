package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"compass/grpc"
	"compass/k8s/portforward"
	"compass/k8s/tunnel"
	"compass/version"

	needlepb "compass/proto/needle/v1"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	survey "gopkg.in/AlecAivazis/survey.v1"
)

var (
	// tunnel to needle
	needleTunnel *tunnel.Tunnel
	// needle cient
	client needlepb.NeedleServiceClient
)

// Common Logger
var log = zerolog.New(os.Stdout).With().Fields(map[string]interface{}{
	"version": version.Version(),
	"commit":  version.Commit(),
}).Timestamp().Logger()

// Application entry point
func main() {
	compassCmd().Execute()
}

// setup opens a port forward to needle
func setup(cmd *cobra.Command, _ []string) error {
	readConfig()
	switch viper.GetString(logFormatKey) {
	case "discard":
		log = log.Output(ioutil.Discard)
	case "console":
		log = log.Output(zerolog.ConsoleWriter{
			Out: os.Stdout,
		})
	}
	t, err := portforward.New(portforward.Options{
		Namespace: viper.GetString(kubeNamespaceKey),
		Context:   viper.GetString(kubeContextKey),
	})
	if err != nil {
		return err
	}
	needleTunnel = t
	c, err := grpc.NewClient(t.LocalAddress())
	if err != nil {
		return err
	}
	client = c
	return nil
}

// teardown closes the port foreard to needle if open
func teardown(cmd *cobra.Command, _ []string) error {
	if needleTunnel != nil {
		needleTunnel.Close()
	}
	return nil
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
	pflags.String("config-path", "", "Path to configuration file")
	pflags.String("log-format", "", "Log output format [console|json]")
	// Bind persistent flags
	viper.BindPFlag(configPathKey, pflags.Lookup("config-path"))
	viper.BindPFlag(logFormatKey, pflags.Lookup("log-format"))
	// Add sub commands
	cmd.AddCommand(
		installCmd(),
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
		Use:      "manage",
		Short:    "Add / Update services compass will manage.",
		PreRunE:  setup,
		PostRunE: teardown,
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
	rsp, err := client.PutService(
		context.Background(),
		&needlepb.PutServiceRequest{
			Service: &needlepb.Service{
				LogicalName: ln,
				Namespace:   ns,
				Description: dsc,
			},
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	svc := rsp.GetService()
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
	cmd.AddCommand(dentryListCmd(), dentryPutCmd(), deleteDentryCmd())
	return cmd
}

func dentryListCmd() *cobra.Command {
	var dtab string
	cmd := &cobra.Command{
		Use:      "list",
		Short:    "List dentries for a given delegation table.",
		PreRunE:  setup,
		PostRunE: teardown,
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(dentryList(dtab))
		},
	}
	cmd.Flags().StringVarP(&dtab, "dtab", "d", "", "Delegation table the dentry should be in.")
	return cmd
}

func dentryList(dtab string) int {
	rsp, err := client.Dentries(
		context.Background(),
		&needlepb.DentriesRequest{
			Dtab: dtab,
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	var dentries = rsp.GetDentries()
	fmt.Println(fmt.Sprintf("Dentries for %s", dtab))
	fmt.Println("-----------")
	for i, dentry := range dentries {
		fmt.Println(dentry.GetId())
		fmt.Println(fmt.Sprintf("%s => %s",
			dentry.GetPrefix(),
			dentry.GetDestination()))
		if i != len(dentries)-1 {
			fmt.Println("-----------")
		}
	}
	return 0
}

func dentryPutCmd() *cobra.Command {
	var dtab string
	var prefix string
	var destination string
	var priority int32
	cmd := &cobra.Command{
		Use:      "put",
		Short:    "Add / Update manual dtab dentries.",
		PreRunE:  setup,
		PostRunE: teardown,
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

func putDentry(dtab, prefix, dst string, priority int32) int {
	rsp, err := client.PutDentry(
		context.Background(),
		&needlepb.PutDentryRequest{
			Dentry: &needlepb.Dentry{
				Dtab:        dtab,
				Prefix:      prefix,
				Destination: dst,
				Priority:    priority,
			},
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	var dentry = rsp.GetDentry()
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
		Use:      "delete",
		Short:    "Delete dentry by Id or Dtab & Prefix",
		PreRunE:  setup,
		PostRunE: teardown,
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
	var ctx = context.Background()
	// Get delegation tables
	delegationTablesRsp, err := client.DelegationTables(ctx, &needlepb.DelegationTablesRequest{})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	var dtabNames []string
	for _, dtab := range delegationTablesRsp.GetDelegationTables() {
		dtabNames = append(dtabNames, dtab.GetName())
	}
	// Prompt user to choose delegation table
	var dtab string
	dtabPrompt := &survey.Select{
		Message: "Please select a Delegation Table:",
		Options: dtabNames,
	}
	survey.AskOne(dtabPrompt, &dtab, nil)
	// Get dentries
	dentriesRsp, err := client.Dentries(
		context.Background(),
		&needlepb.DentriesRequest{
			Dtab: dtab,
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	// Prompt user to pick a dentry
	var dentryNames []string
	var dentryMap map[string]*needlepb.Dentry = map[string]*needlepb.Dentry{}
	for _, dentry := range dentriesRsp.GetDentries() {
		name := fmt.Sprintf("%s => %s", dentry.GetPrefix(), dentry.GetDestination())
		dentryNames = append(dentryNames, name)
		dentryMap[name] = dentry
	}
	var dentryName string
	dentryPrompt := &survey.Select{
		Message: "Please select a Dentry to delete:",
		Options: dentryNames,
	}
	survey.AskOne(dentryPrompt, &dentryName, nil)
	// Confirm prompt
	dentry, ok := dentryMap[dentryName]
	if !ok {
		fmt.Println("Dentry not found")
		return 1
	}
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("Delete %s => %s?", dentry.GetPrefix(), dentry.GetDestination()),
	}
	survey.AskOne(confirmPrompt, &confirm, nil)
	if !confirm {
		fmt.Println("Dentry was not deleted.")
		return 0
	}
	// Delete dentry
	deleteDentryByIdRsp, err := client.DeleteDentryById(
		context.Background(),
		&needlepb.DeleteDentryByIdRequest{
			Id: dentry.GetId(),
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if !deleteDentryByIdRsp.GetDeleted() {
		fmt.Println("Dentry was not deleted.")
		return 1
	}
	fmt.Println("Dentry has been deleted")
	return 0
}

func deleteDentryById(id string) int {
	rsp, err := client.DeleteDentryById(
		context.Background(),
		&needlepb.DeleteDentryByIdRequest{
			Id: id,
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if !rsp.GetDeleted() {
		fmt.Println("Dentry was not deleted")
		return 1
	}
	fmt.Println("Dentry has been deleted")
	return 0
}

func deleteDentryByPrefix(dtab, prefix string) int {
	rsp, err := client.DeleteDentryByPrefix(
		context.Background(),
		&needlepb.DeleteDentryByPrefixRequest{
			Dtab:   dtab,
			Prefix: prefix,
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if !rsp.GetDeleted() {
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
		Use:      "version",
		Short:    "Route a service a specifcly deployed version",
		PreRunE:  setup,
		PostRunE: teardown,
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(routeVersion(logicalName, version))
		},
	}
	cmd.Flags().StringVarP(&logicalName, "logical-name", "l", "", "Service logical name.")
	cmd.Flags().StringVarP(&version, "version", "v", "", "Version e.g: develop, 1.2.3.")
	return cmd
}

func routeVersion(logicalName, version string) int {
	_, err := client.RouteToVersion(
		context.Background(),
		&needlepb.RouteToVersionRequest{
			LogicalName: logicalName,
			Version:     version,
		})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println(fmt.Sprintf("Updated delegation table to route %s to %s", logicalName, version))
	return 0
}

package main

import (
	"fmt"
	"os"

	"compass/k8s"
	"compass/needle/install"

	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	var namespace string
	var rbac bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install needle, compass's server component into a Kubernetes cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			os.Exit(installCompass(
				namespace,
				rbac,
			))
		},
	}
	cmd.Flags().BoolVar(&rbac, "with-rbac", false, "Create RBAC resources")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to install Needle into")
	return cmd
}

func installCompass(namespace string, rbac bool) int {
	fmt.Println("Installing needle, the compass server into the cluster...")
	client, err := k8s.Clientset()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	var opts []install.Option
	if namespace != "" {
		opts = append(opts, install.WithNamespace(namespace))
	}
	if rbac {
		opts = append(opts, install.WithRBAC())
	}
	if err := install.Install(client, opts...); err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println("Install complete")
	return 0
}

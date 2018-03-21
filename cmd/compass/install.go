package main

import (
	"fmt"
	"os"

	"compass/pkg/kube"

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
	client, err := kube.Clientset()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	var opts []kube.InstallOption
	if namespace != "" {
		opts = append(opts, kube.WithInstallNamespace(namespace))
	}
	if rbac {
		opts = append(opts, kube.WithInstallRBAC())
	}
	if err := kube.Install(client, opts...); err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Println("Install complete")
	return 0
}

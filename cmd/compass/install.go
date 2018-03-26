package main

import (
	"compass/pkg/kube"
	"compass/pkg/ui"

	"github.com/spf13/cobra"
)

// installCmd returns a cobra command for installing compass into a kubernetes
// existing cluster
func installCmd() *cobra.Command {
	var namespace string
	var rbac bool
	var namerd string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install needle, compass's server component into a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := kube.Clientset()
			if err != nil {
				return err
			}
			if namespace == "" || namerd == "" {
				return ui.Installer(clientset)
			}
			var opts = []kube.InstallOption{
				kube.InstallWithNamespace(namespace),
				kube.InstallWithNamerdHost(namerd),
			}
			if rbac {
				opts = append(opts, kube.InstallWithRBAC())
			}
			return kube.Install(clientset, opts...)
		},
	}
	cmd.Flags().BoolVar(&rbac, "with-rbac", false, "Create RBAC resources")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to install Needle into")
	cmd.Flags().StringVar(&namerd, "namerd", "", "Namerd HTTP API host:port, e.g: namerd.namerd.svc.cluster.local:4180")
	return cmd
}

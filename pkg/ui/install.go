package ui

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"compass/pkg/kube"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var installIntroMsg = asciiLogo + `
This tool guide you through the process of installing and configuring Compass.
Compass requires namerd to be installed in your Cluster with a service exposing
the namerd HTTP API. See the namerd documentation for help:

https://linkerd.io/config/latest/namerd/index.html#http-controller`

var installSummaryMsg = `
Compass will be installed with the following settings:
Namespace: {{.Namespace}}
With RBAC: {{if .RBAC}}yes{{else}}no{{end}}
Namerd Host: {{.NamerdHost}}`

// Compiled templates
var (
	installIntroTpl   = template.Must(template.New("installIntroMsg").Parse(installIntroMsg))
	installSummaryTpl = template.Must(template.New("installSummaryMsg").Parse(installSummaryMsg))
)

// Errors
var (
	ErrNotReady              = errors.New("not ready")
	ErrNamerdServiceNotFound = errors.New("could not find namerd service")
	ErrNotInstalled          = errors.New("not installed")
)

// InstallSummaryData is the data passed to the install summary template
type InstallSummaryData struct {
	Namespace  string
	RBAC       bool
	NamerdHost string
}

// Installer runs an interactive user driven install wizard
func Installer(cs *kubernetes.Clientset) error {
	if err := installIntro(); err != nil {
		return err
	}
	if !Confirm(ConfirmMessage("Ready?"), ConfirmDefault(true)) {
		return ErrNotReady
	}
	var opts []kube.InstallOption
	var rbac bool
	if Confirm(ConfirmMessage("Intall RBAC roles?"), ConfirmDefault(true)) {
		rbac = true
		opts = append(opts, kube.InstallWithRBAC())
	}
	namespace, err := installNamespace(cs)
	if err != nil {
		return err
	}
	opts = append(opts, kube.InstallWithNamespace(namespace.GetName()))
	namerdHost, err := installNamerdHost(cs)
	if err != nil {
		return err
	}
	opts = append(opts, kube.InstallWithNamerdHost(namerdHost))
	installSummary(InstallSummaryData{
		Namespace:  namespace.GetName(),
		RBAC:       rbac,
		NamerdHost: namerdHost,
	})
	if !Confirm(ConfirmMessage("Install Compass?"), ConfirmDefault(true)) {
		return ErrNotInstalled
	}
	return kube.Install(cs, opts...)
}

// installIntro prints the install intro message
func installIntro() error {
	var w bytes.Buffer
	if err := installIntroTpl.Execute(&w, nil); err != nil {
		return err
	}
	fmt.Println(w.String())
	return nil
}

// installSummary prints the install summary
func installSummary(data InstallSummaryData) error {
	var w bytes.Buffer
	if err := installSummaryTpl.Execute(&w, &data); err != nil {
		return err
	}
	fmt.Println(w.String())
	return nil
}

// installNamespace asks the user to confirm the install namespace
func installNamespace(cs *kubernetes.Clientset) (*corev1.Namespace, error) {
	ns, err := cs.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var choices []string
	for _, n := range ns.Items {
		choices = append(choices, n.GetName())
	}
	i, _ := Choice(
		choices,
		ChoiceMessage("Select a namespace to install Compass in:"),
		ChoiceDefault("kube-system"))
	return &ns.Items[i], nil
}

// installNamerdHost verifies the namerd host
func installNamerdHost(cs *kubernetes.Clientset) (string, error) {
	var host string
	if Confirm(ConfirmMessage("Auto discover namerd serivce?"), ConfirmDefault(true)) {
		port := Input(
			InputMessage("What port does the Namerd HTTP API Port run on?"),
			InputDefault("4180"))
		for {
			service, err := discoverNamerdService(cs)
			if err != nil {
				return "", err
			}
			host = fmt.Sprintf(
				"%s.%s.svc.cluster.local:%s",
				service.GetName(),
				service.GetNamespace(),
				port)
			if Confirm(ConfirmMessage(fmt.Sprintf("Use %s as namerd host?", host)), ConfirmDefault(true)) {
				break
			}
		}
	} else {
		host = Input(InputMessage("Namerd Host (host:port):"))
	}
	return host, nil
}

// discoverNamerdService discovers the namerd service based on labels
func discoverNamerdService(cs *kubernetes.Clientset) (*corev1.Service, error) {
	namespaces, err := cs.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	labels := labels.Set{"app": "namerd"}
	for {
		for _, n := range namespaces.Items {
			fmt.Println(fmt.Sprintf("Searching for namerd service with labels: %s", labels.AsSelector().String()))
			sl, err := cs.CoreV1().Services(n.Namespace).List(metav1.ListOptions{
				LabelSelector: labels.AsSelector().String(),
			})
			if err != nil {
				return nil, err
			}
			if len(sl.Items) == 1 {
				return &sl.Items[0], nil
			}
			fmt.Println("Could not find namerd service")
			if !Confirm(ConfirmMessage("Retry with different labels?"), ConfirmDefault(true)) {
				return nil, ErrNamerdServiceNotFound
			}
			labels = LabelsPrompt()
		}
	}
}

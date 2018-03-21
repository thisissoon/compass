package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Allows kubernetes config overrides
type OverrideOption func(o *clientcmd.ConfigOverrides)

// WithCurrentContext overrides a kubernetes client context
func WithCurrentContext(c string) OverrideOption {
	return func(o *clientcmd.ConfigOverrides) {
		o.CurrentContext = c
	}
}

// Returns a kubernetes client config - with no overrides
func Config(overrides ...OverrideOption) clientcmd.ClientConfig {
	lr := clientcmd.NewDefaultClientConfigLoadingRules()
	co := &clientcmd.ConfigOverrides{}
	for _, override := range overrides {
		override(co)
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(lr, co)
}

// RestClient creates a Kubernetes config and client for a given kubeconfig context.
func RestClient(overrides ...OverrideOption) (*rest.Config, kubernetes.Interface, error) {
	config, err := Config(overrides...).ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

// Clientset creates a new kubernetes Clientset
func Clientset(overrides ...OverrideOption) (*kubernetes.Clientset, error) {
	config := Config(overrides...)
	c, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c)
}

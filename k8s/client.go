package k8s

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	// kubernetes auth plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Selectors allow us to build selector labels
type Selectors map[string]string

// Add adds a new selector
func (s Selectors) Add(label, value string) {
	s[label] = value
}

// Del removes a selector
func (s Selectors) Del(label string) {
	delete(s, label)
}

// String returns the selectors in a string format: foo=bar,fizz=buzz
func (s Selectors) String() string {
	var i int
	selectors := make([]string, len(s))
	for k, v := range s {
		selectors[i] = fmt.Sprintf("%s=%s", k, v)
		i += 1
	}
	return strings.Join(selectors, ",")
}

// An Option function allows for Client configuration overrides
type Option func(*Client)

// WithNamespace overrides the Client default namespace
func WithNamespace(ns string) Option {
	return func(c *Client) {
		c.namespace = ns
	}
}

// Client provides wraps a kubernetes.Clientset providing
// convienience methods
type Client struct {
	namespace string
	client    *kubernetes.Clientset
}

// ListServices returns a list a kubernetes services that match the
// given selectors
func (c *Client) ListServices(s Selectors) ([]v1.Service, error) {
	list, err := c.client.Core().Services(c.namespace).List(meta.ListOptions{
		LabelSelector: s.String(),
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// New constructs a new Client
func New(client *kubernetes.Clientset, opts ...Option) *Client {
	c := &Client{
		namespace: "default",
		client:    client,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Clientset creates a new kubernetes Clientset
func Clientset() (*kubernetes.Clientset, error) {
	lr := clientcmd.NewDefaultClientConfigLoadingRules()
	co := &clientcmd.ConfigOverrides{}
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(lr, co)
	c, err := cc.ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(c)
}

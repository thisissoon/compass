package kube

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// An Option configures a Tunnel
type TunnelOption func(o *TunnelOptions)

// WithPodName returns an Option for configuring the namespace
func TunnelWithNamespace(namespace string) TunnelOption {
	return func(o *TunnelOptions) {
		o.Namespace = namespace
	}
}

// TunnelWithPodLabels sets the pod labels to tunnel too
func TunnelWithPodLabels(labels labels.Set) TunnelOption {
	return func(o *TunnelOptions) {
		o.PodLabels = labels
	}
}

// WithRemotePort returns an Option for configuring the Tunnel remote port
func TunnelWithRemotePort(port int) TunnelOption {
	return func(o *TunnelOptions) {
		o.RemotePort = port
	}
}

// Default Tunnel Options
var DefaultTunnelOptions = TunnelOptions{
	Namespace: "kube-system",
	PodLabels: labels.Set{
		"name": "needle",
		"app":  "compass",
	},
	RemotePort: 5000,
}

// TunnelOptions configures a Tunnel
type TunnelOptions struct {
	Namespace  string
	PodLabels  labels.Set
	RemotePort int
}

// Tunnel opens a tunnel to a kubernetes port locally
type Tunnel struct {
	// Options to configure the tunnel
	Options TunnelOptions

	// Local port the port forward is bound too
	localPort int

	// Channels to manage the port forward
	stopCh  chan struct{}
	readyCh chan struct{}
}

// New constructs a new Tunnel
func NewTunnel(opts ...TunnelOption) *Tunnel {
	var options = DefaultTunnelOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &Tunnel{
		Options: options,
		stopCh:  make(chan struct{}, 1),
		readyCh: make(chan struct{}, 1),
	}
}

// Port returns the local port the port forward is served on
func (t *Tunnel) Port() int {
	return t.localPort
}

// LocalAddress returns the local address, e.g 127.0.0.1:12345
func (t *Tunnel) LocalAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", t.Port())
}

// Open opens the tunnel
func (t *Tunnel) Open() error {
	// Get a kubernetets rest client
	config, client, err := RestClient()
	if err != nil {
		return err
	}
	// Get needle pod
	pod, err := getFirstRunningPod(
		client.CoreV1(),
		t.Options.Namespace,
		t.Options.PodLabels)
	if err != nil {
		return err
	}
	// Get pod url
	purl := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(t.Options.Namespace).
		Name(pod).
		SubResource("portforward").URL()
	// Create a spdy transport / upgrader from config
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}
	// Create pod dialer
	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		"POST",
		purl)
	// Get a free local port
	t.localPort, err = FreePort()
	if err != nil {
		return err
	}
	// Create a new port forwarder
	ports := []string{fmt.Sprintf("%d:%d", t.localPort, t.Options.RemotePort)}
	forwarder, err := portforward.New(
		dialer,
		ports,
		t.stopCh,
		t.readyCh,
		ioutil.Discard,
		ioutil.Discard)
	if err != nil {
		return err
	}
	// Start the port forwarder - capturing errors to an error channel
	errCh := make(chan error)
	go func() {
		errCh <- forwarder.ForwardPorts()
	}()
	// Block until error or ready
	select {
	case err = <-errCh:
		return fmt.Errorf("forwarding ports: %v", err)
	case <-forwarder.Ready:
		return nil
	}
}

// Close stops the port forward
func (t *Tunnel) Close() {
	close(t.stopCh)
}

// FreePort returns a free open port to forward requests too locally
func FreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, err
}

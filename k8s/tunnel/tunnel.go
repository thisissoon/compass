package tunnel

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// An Option configures a Tunnel
type Option func(t *Tunnel)

// WithPodName returns an Option for configuring the namespace
func WithNamespace(namespace string) Option {
	return func(t *Tunnel) {
		t.Namespace = namespace
	}
}

// WithPodName returns an Option for configuring the pod name
func WithPodName(pod string) Option {
	return func(t *Tunnel) {
		t.PodName = pod
	}
}

// WithRemotePort returns an Option for configuring the Tunnel remote port
func WithRemotePort(port int) Option {
	return func(t *Tunnel) {
		t.RemotePort = port
	}
}

// Tunnel opens a tunnel to a kubernetes port locally
type Tunnel struct {
	Namespace  string
	PodName    string
	RemotePort int

	client    rest.Interface
	config    *rest.Config
	localPort int

	stopCh  chan struct{}
	readyCh chan struct{}
}

// New constructs a new Tunnel
func New(client rest.Interface, config *rest.Config, opts ...Option) *Tunnel {
	t := &Tunnel{
		Namespace:  "default",
		PodName:    "needle",
		RemotePort: 5000,

		client: client,
		config: config,

		stopCh:  make(chan struct{}, 1),
		readyCh: make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
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
	// Get pod url
	u := t.client.Post().
		Resource("pods").
		Namespace(t.Namespace).
		Name(t.PodName).
		SubResource("portforward").URL()
	// Create a spdy transport / upgrader from config
	transport, upgrader, err := spdy.RoundTripperFor(t.config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", u)
	// Get a free local port
	t.localPort, err = FreePort()
	if err != nil {
		return err
	}
	// Create a new port forwarder
	ports := []string{fmt.Sprintf("%d:%d", t.localPort, t.RemotePort)}
	pf, err := portforward.New(dialer, ports, t.stopCh, t.readyCh, ioutil.Discard, ioutil.Discard)
	if err != nil {
		return err
	}
	// Start the port forwarder - capturing errors to an error channel
	errCh := make(chan error)
	go func() {
		errCh <- pf.ForwardPorts()
	}()
	// Block until error or ready
	select {
	case err = <-errCh:
		return fmt.Errorf("forwarding ports: %v", err)
	case <-pf.Ready:
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

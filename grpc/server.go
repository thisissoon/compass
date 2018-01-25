package grpc

import (
	"net"
	"sync"

	"google.golang.org/grpc"

	needle "compass/proto/needle/v1"
)

// An Option funtion can override configuration options
// for a server
type Option func(*Server)

// WithAddress overrides the default configured listen
// address for a server
func WithAddress(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// A Server can create and stop a gRPC server
type Server struct {
	addr string                      // address to bind too
	sm   needle.ServiceManagerServer // service manager

	lock sync.Mutex // protects below
	srv  *grpc.Server
}

// Start starts serving the gRPC server
func (s *Server) Serve() (net.Addr, <-chan error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	errC := make(chan error, 1)
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		errC <- err
		return nil, (<-chan error)(errC)
	}
	s.srv = grpc.NewServer()
	needle.RegisterServiceManagerServer(s.srv, s.sm)
	go func() { // Start serving :)
		errC <- s.srv.Serve(ln)
	}()
	return ln.Addr(), (<-chan error)(errC)
}

// Stop gracefully stops the grpc server
func (s *Server) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.srv != nil {
		s.srv.GracefulStop()
	}
}

// NewServer creates a new gRPC server
// Use Option functions to override defaults
func NewServer(sm needle.ServiceManagerServer, opts ...Option) *Server {
	s := &Server{
		addr: ListenAddress(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

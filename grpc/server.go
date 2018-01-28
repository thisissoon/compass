package grpc

import (
	"net"
	"sync"

	"google.golang.org/grpc"

	"compass/logger"
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
	addr string                     // address to bind too
	ns   needle.NeedleServiceServer // needle service service

	lock sync.Mutex // protects below
	srv  *grpc.Server
}

// Start starts serving the gRPC server
func (s *Server) Serve() <-chan error {
	s.lock.Lock()
	defer s.lock.Unlock()
	log := logger.New()
	errC := make(chan error, 1)
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		errC <- err
		return (<-chan error)(errC)
	}
	s.srv = grpc.NewServer()
	needle.RegisterNeedleServiceServer(s.srv, s.ns)
	go func() { // Start serving :)
		log.Debug().Str("address", ln.Addr().String()).Msg("gRPC server started")
		errC <- s.srv.Serve(ln)
	}()
	return (<-chan error)(errC)
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
func NewServer(ns needle.NeedleServiceServer, opts ...Option) *Server {
	s := &Server{
		addr: ListenAddress(),
		ns:   ns,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

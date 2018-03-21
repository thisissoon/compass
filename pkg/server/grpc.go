package server

import (
	"net"
	"sync"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	pb "compass/pkg/proto/services"
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
	addr string
	svc  pb.DentryServiceServer

	lock sync.Mutex // protects below
	srv  *grpc.Server
}

// Start starts serving the gRPC server
func (s *Server) Serve(log zerolog.Logger) <-chan error {
	s.lock.Lock()
	defer s.lock.Unlock()
	errC := make(chan error, 1)
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		errC <- err
		return (<-chan error)(errC)
	}
	s.srv = grpc.NewServer(
		grpc.UnaryInterceptor(LogUnaryInterceptor(log)),
		grpc.StreamInterceptor(LogStreamInyerceptor(log)))
	pb.RegisterDentryServiceServer(s.srv, s.svc)
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
func New(svc pb.DentryServiceServer, opts ...Option) *Server {
	s := &Server{
		addr: ":5000",
		svc:  svc,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

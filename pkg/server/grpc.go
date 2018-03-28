package server

import (
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// Options configures the gRPC server
type Options struct {
	Address string
	Log     zerolog.Logger
}

// An Option funtion can override configuration options
// for a server
type Option func(*Options)

// WithAddress overrides the default configured listen address for a server
func WithAddress(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}

// WithLogger overrides the default logger for a server
func WithLogger(logger zerolog.Logger) Option {
	return func(o *Options) {
		o.Log = logger
	}
}

// A Service can register itself with a server
type Service interface {
	RegisterWithServer(*grpc.Server)
}

// A Server can create and stop a gRPC server
type Server struct {
	Options Options

	server *grpc.Server
}

// Start starts serving the gRPC server
func (s *Server) Serve(services ...Service) error {
	ln, err := net.Listen("tcp", s.Options.Address)
	if err != nil {
		return err
	}
	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(LogUnaryInterceptor(s.Options.Log)),
		grpc.StreamInterceptor(LogStreamInyerceptor(s.Options.Log)))
	for _, service := range services {
		service.RegisterWithServer(s.server)
	}
	return s.server.Serve(ln)
}

// Stop gracefully stops the grpc server
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// NewServer creates a new gRPC server
// Use Option functions to override defaults
func New(opts ...Option) *Server {
	options := Options{
		Address: ":5000",
		Log:     zerolog.Nop(),
	}
	s := &Server{
		Options: options,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return s
}

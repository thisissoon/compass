package service

import (
	"context"

	"compass/logger"
	"compass/store"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	needle "compass/proto/needle/v1"
)

// Manager implements the needle.ServiceManagerServer interface
// and thus handles gRPC requests
type Manager struct {
	store store.ServiceStore
}

// PutService creates or updates a service
func (m *Manager) PutService(ctx context.Context, req *needle.PutServiceRequest) (*needle.PutServiceResponse, error) {
	return putService(m.store, req.GetService(), logger.New())
}

// putService is the underlying logic for handling a gRPC put service request
func putService(sp store.ServicePutter, s *needle.Service, log zerolog.Logger) (*needle.PutServiceResponse, error) {
	svc, err := sp.PutService(&store.Service{
		LogicalName: s.GetLogicalName(),
		Namespace:   s.GetNamespace(),
		Description: s.GetDescription(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to put service")
		return nil, status.Error(codes.Internal, "unable to put service")
	}
	return &needle.PutServiceResponse{
		Service: &needle.Service{
			LogicalName: svc.LogicalName,
			Namespace:   svc.Namespace,
			Description: svc.Description,
		},
	}, nil
}

// NewManager returns a new Manager
func NewManager(store store.ServiceStore) *Manager {
	return &Manager{
		store: store,
	}
}

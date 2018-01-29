package needle

import (
	"context"

	"compass/logger"
	"compass/store"

	needle "compass/proto/needle/v1"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	uuid "github.com/satori/go.uuid"
)

// Manager implements the needle.NeedleServiceServer interface
// and thus handles gRPC requests
type Service struct {
	store store.Store
}

// PutService creates or updates a service
func (s *Service) PutService(ctx context.Context, req *needle.PutServiceRequest) (*needle.PutServiceResponse, error) {
	return putService(s.store, req.GetService(), logger.New())
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
			Id:          svc.Id.String(),
			CreateDate:  &timestamp.Timestamp{Seconds: svc.CreateDate.Unix()},
			UpdateDate:  &timestamp.Timestamp{Seconds: svc.UpdateDate.Unix()},
			LogicalName: svc.LogicalName,
			Namespace:   svc.Namespace,
			Description: svc.Description,
		},
	}, nil
}

// PutDentry creates or updates a dentry
func (s *Service) PutDentry(ctx context.Context, req *needle.PutDentryRequest) (*needle.PutDentryResponse, error) {
	return putDentry(s.store, req.GetDentry(), logger.New())
}

// putDentry is the underlying logic for handling a gRPC put dentry request
func putDentry(sp store.DentryPutter, d *needle.Dentry, log zerolog.Logger) (*needle.PutDentryResponse, error) {
	dentry, err := sp.PutDentry(&store.Dentry{
		Id:          uuid.FromStringOrNil(d.GetId()),
		Dtab:        d.GetDtab(),
		Prefix:      d.GetPrefix(),
		Destination: d.GetDestination(),
		Priority:    d.GetPriority(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to put dentry")
		return nil, status.Error(codes.Internal, "unable to put dentry")
	}
	return &needle.PutDentryResponse{
		Dentry: &needle.Dentry{
			Id:          dentry.Id.String(),
			CreateDate:  &timestamp.Timestamp{Seconds: dentry.CreateDate.Unix()},
			UpdateDate:  &timestamp.Timestamp{Seconds: dentry.UpdateDate.Unix()},
			Dtab:        dentry.Dtab,
			Prefix:      dentry.Prefix,
			Destination: dentry.Destination,
			Priority:    dentry.Priority,
		},
	}, nil
}

func (s *Service) RouteToVersion(ctx context.Context, req *needle.RouteToVersionRequest) (*needle.RouteToVersionResponse, error) {
	return nil, nil
}

// NewService returns a new Manager
func NewService(store store.Store) *Service {
	return &Service{
		store: store,
	}
}

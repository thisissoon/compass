package needle

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"compass/k8s"
	"compass/logger"
	"compass/store"

	needle "compass/proto/needle/v1"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	uuid "github.com/satori/go.uuid"
)

const (
	DtabAnnotation           = "compass.thisissoon.com/dtab"
	DentryPrefixAnnotation   = "compass.thisissoon.com/dentry-prefix"
	DentryPriorityAnnotation = "compass.thisissoon.com/dentry-priority"
	PortNameAnnotation       = "compass.thisissoon.com/port-name"
)

type ServiceSelectorDentryPutter interface {
	store.ServiceSelector
	store.DentryPutter
}

// Manager implements the needle.NeedleServiceServer interface
// and thus handles gRPC requests
type Service struct {
	store store.Store
	k8s   *k8s.Client
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

// RouteToVersion routes a service to a specified version
func (s *Service) RouteToVersion(ctx context.Context, req *needle.RouteToVersionRequest) (*needle.RouteToVersionResponse, error) {
	return routeToVersion(s.store, s.k8s, req.GetLogicalName(), req.GetVersion())
}

// routeToVersion queries kubernetets for service matching the provided
// logical name and version.
// If a valid service is found with the required annotations then a dentry is
// upserted and a dtab sync triggered to update the namerd dtab for the service
func routeToVersion(s ServiceSelectorDentryPutter, sl k8s.ServiceLister, ln, ver string) (*needle.RouteToVersionResponse, error) {
	// Get service by it's logical name
	svc, err := s.GetByLogicalName(ln)
	if err != nil {
		return nil, err
	}
	// Find coresponding service by service logical name and the version
	kServices, err := sl.ListServices(
		svc.Namespace,
		k8s.Selectors{
			"logicalName": svc.LogicalName,
			"version":     ver,
		})
	if err != nil {
		return nil, err
	}
	// Ensure we have only 1 service
	if len(kServices) != 1 {
		return nil, fmt.Errorf("expected 1 kubernetes service, found: %d", len(kServices))
	}
	ks := kServices[0]
	// Lookup the dtab annotation so we know what dtab the dentry is managed within
	dtab, ok := ks.Annotations[DtabAnnotation]
	if !ok {
		return nil, fmt.Errorf("missing required annotation: %s", DtabAnnotation)
	}
	// Lookup the port-name annotation so we know what port-name to store in the dentry
	portName, ok := ks.Annotations[PortNameAnnotation]
	if !ok {
		return nil, fmt.Errorf("missing required annotation: %s", PortNameAnnotation)
	}
	// Lookup the prefix annotation with a default prefix being the service logical name
	prefix := fmt.Sprintf("/%s", svc.LogicalName)
	if prefixAnnotation, ok := ks.Annotations[DentryPrefixAnnotation]; ok {
		prefix = prefixAnnotation
	}
	// Lookup priority annotation so we we can place the dentry in the appropriate place in teh dtab
	priorityStr, ok := ks.Annotations[DentryPriorityAnnotation]
	if ok {
		return nil, fmt.Errorf("missing required annotation: %s", DentryPriorityAnnotation)
	}
	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		return nil, errors.New("could not convert prioroty to int")
	}
	// Put the dentry to storage
	_, err = s.PutDentry(&store.Dentry{
		Prefix:      prefix,
		Destination: fmt.Sprintf("/#/io.l5d.k8s/%s/%s/%s", ks.Namespace, portName, ks.Name),
		Dtab:        dtab,
		Priority:    int32(priority),
	})
	if err != nil {
		return nil, err
	}
	// Return response
	return &needle.RouteToVersionResponse{}, nil
}

// NewService returns a new Manager
func NewService(store store.Store, k8s *k8s.Client) *Service {
	return &Service{
		store: store,
		k8s:   k8s,
	}
}

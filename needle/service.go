package needle

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"compass/k8s"
	"compass/store"

	needle "compass/proto/needle/v1"

	"github.com/golang/protobuf/ptypes/timestamp"
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
	return putService(ctx, s.store, req.GetService())
}

// putService is the underlying logic for handling a gRPC put service request
func putService(ctx context.Context, sp store.ServicePutter, s *needle.Service) (*needle.PutServiceResponse, error) {
	svc, err := sp.PutService(&store.Service{
		LogicalName: s.GetLogicalName(),
		Namespace:   s.GetNamespace(),
		Description: s.GetDescription(),
	})
	if err != nil {
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
	return putDentry(ctx, s.store, req.GetDentry())
}

// putDentry is the underlying logic for handling a gRPC put dentry request
func putDentry(ctx context.Context, sp store.DentryPutter, d *needle.Dentry) (*needle.PutDentryResponse, error) {
	dentry, err := sp.PutDentry(&store.Dentry{
		Id:          uuid.FromStringOrNil(d.GetId()),
		Dtab:        d.GetDtab(),
		Prefix:      d.GetPrefix(),
		Destination: d.GetDestination(),
		Priority:    d.GetPriority(),
	})
	if err != nil {
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

// DeleteDentryById deletes a dentry by Id
func (s *Service) DeleteDentryById(ctx context.Context, req *needle.DeleteDentryByIdRequest) (*needle.DeleteDentryByIdResponse, error) {
	return deleteDentryById(s.store, req.GetId())
}

// deleteDentryById deletes a dentry by Id
func deleteDentryById(db store.DentryByIdDeletor, id string) (*needle.DeleteDentryByIdResponse, error) {
	affected, err := db.DeleteDentryById(uuid.FromStringOrNil(id))
	if err != nil {
		return nil, err
	}
	return &needle.DeleteDentryByIdResponse{
		Deleted: (affected > 0),
	}, nil
}

// DeleteDentryByPrefix deletes dentry by prefix
func (s *Service) DeleteDentryByPrefix(ctx context.Context, req *needle.DeleteDentryByPrefixRequest) (*needle.DeleteDentryByPrefixResponse, error) {
	return deleteDentryByPrefix(s.store, req.GetDtab(), req.GetPrefix())
}

// deleteDentryByPrefix deletes dentry by prefix
func deleteDentryByPrefix(db store.DentryByPrefixDeletor, dtab, prefix string) (*needle.DeleteDentryByPrefixResponse, error) {
	affected, err := db.DeleteDentryByPrefix(dtab, prefix)
	if err != nil {
		return nil, err
	}
	return &needle.DeleteDentryByPrefixResponse{
		Deleted: (affected > 0),
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
	if !ok {
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

// DelegationTables returns a list of delgation tables managed by needle
func (n *Service) DelegationTables(ctx context.Context, req *needle.DelegationTablesRequest) (*needle.DelegationTablesResponse, error) {
	tables, err := n.store.DelegationTables()
	if err != nil {
		return nil, err
	}
	var dtabs []*needle.DelegationTable
	for _, t := range tables {
		dtabs = append(dtabs, &needle.DelegationTable{
			Name: t,
		})
	}
	return &needle.DelegationTablesResponse{
		DelegationTables: dtabs,
	}, nil
}

// Dentries returns a list of dentries for a delegation table
func (n *Service) Dentries(ctx context.Context, req *needle.DentriesRequest) (*needle.DentriesResponse, error) {
	dCh, err := n.store.DentriesByDtab(ctx, req.GetDtab())
	if err != nil {
		return nil, err
	}
	var dentries []*needle.Dentry
	for dentry := range dCh {
		dentries = append(dentries, &needle.Dentry{
			Id:          dentry.Id.String(),
			CreateDate:  &timestamp.Timestamp{Seconds: dentry.CreateDate.Unix()},
			UpdateDate:  &timestamp.Timestamp{Seconds: dentry.UpdateDate.Unix()},
			Dtab:        dentry.Dtab,
			Prefix:      dentry.Prefix,
			Destination: dentry.Destination,
			Priority:    dentry.Priority,
		})
	}
	return &needle.DentriesResponse{
		Dentries: dentries,
	}, nil
}

// NewService returns a new Manager
func NewService(store store.Store, k8s *k8s.Client) *Service {
	return &Service{
		store: store,
		k8s:   k8s,
	}
}

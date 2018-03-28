package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"compass/pkg/store"

	pb "compass/pkg/proto/services"

	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	uuid "github.com/satori/go.uuid"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	DtabAnnotation           = "compass.thisissoon.com/dtab"
	DentryPrefixAnnotation   = "compass.thisissoon.com/dentry-prefix"
	DentryPriorityAnnotation = "compass.thisissoon.com/dentry-priority"
	PortNameAnnotation       = "compass.thisissoon.com/port-name"
)

var (
	ErrServiceNotFound                 = errors.New("kubernetets service not found")
	ErrMultipleServicesFound           = errors.New("Too many kubernetets services found")
	ErrMissingDtabAnnotation           = fmt.Errorf("service missing annotation: %s", DtabAnnotation)
	ErrMissingDentryPriorityAnnotation = fmt.Errorf("service missing annotation: %s", DentryPriorityAnnotation)
	ErrMissingPortNameAnnotation       = fmt.Errorf("service missing annotation: %s", PortNameAnnotation)
)

type ServiceSelectorDentryPutter interface {
	store.ServiceSelector
	store.DentryPutter
}

// Manager implements the pb.NeedleServiceServer interface
// and thus handles gRPC requests
type DentryService struct {
	store store.Store
	k8s   *kubernetes.Clientset
}

// NewDentryService returns a new DentryService
func NewDentryService(store store.Store, k8s *kubernetes.Clientset) *DentryService {
	return &DentryService{
		store: store,
		k8s:   k8s,
	}
}

// PutService creates or updates a service
func (s *DentryService) PutService(ctx context.Context, req *pb.PutServiceRequest) (*pb.PutServiceResponse, error) {
	return putService(ctx, s.store, req.GetService())
}

// putService is the underlying logic for handling a gRPC put service request
func putService(ctx context.Context, sp store.ServicePutter, s *pb.Service) (*pb.PutServiceResponse, error) {
	svc, err := sp.PutService(&store.Service{
		LogicalName: s.GetLogicalName(),
		Namespace:   s.GetNamespace(),
		Description: s.GetDescription(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "unable to put service")
	}
	return &pb.PutServiceResponse{
		Service: &pb.Service{
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
func (s *DentryService) PutDentry(ctx context.Context, req *pb.PutDentryRequest) (*pb.PutDentryResponse, error) {
	return putDentry(ctx, s.store, req.GetDentry())
}

// putDentry is the underlying logic for handling a gRPC put dentry request
func putDentry(ctx context.Context, sp store.DentryPutter, d *pb.Dentry) (*pb.PutDentryResponse, error) {
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
	return &pb.PutDentryResponse{
		Dentry: &pb.Dentry{
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
func (s *DentryService) DeleteDentryById(ctx context.Context, req *pb.DeleteDentryByIdRequest) (*pb.DeleteDentryByIdResponse, error) {
	return deleteDentryById(s.store, req.GetId())
}

// deleteDentryById deletes a dentry by Id
func deleteDentryById(db store.DentryByIdDeletor, id string) (*pb.DeleteDentryByIdResponse, error) {
	affected, err := db.DeleteDentryById(uuid.FromStringOrNil(id))
	if err != nil {
		return nil, err
	}
	return &pb.DeleteDentryByIdResponse{
		Deleted: (affected > 0),
	}, nil
}

// DeleteDentryByPrefix deletes dentry by prefix
func (s *DentryService) DeleteDentryByPrefix(ctx context.Context, req *pb.DeleteDentryByPrefixRequest) (*pb.DeleteDentryByPrefixResponse, error) {
	return deleteDentryByPrefix(s.store, req.GetDtab(), req.GetPrefix())
}

// deleteDentryByPrefix deletes dentry by prefix
func deleteDentryByPrefix(db store.DentryByPrefixDeletor, dtab, prefix string) (*pb.DeleteDentryByPrefixResponse, error) {
	affected, err := db.DeleteDentryByPrefix(dtab, prefix)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteDentryByPrefixResponse{
		Deleted: (affected > 0),
	}, nil
}

// RouteToVersion routes a service to a specified version
func (s *DentryService) RouteToVersion(ctx context.Context, req *pb.RouteToVersionRequest) (*pb.RouteToVersionResponse, error) {
	return routeToVersion(
		s.store,
		s.k8s,
		req.GetLogicalName(),
		req.GetVersion())
}

func getKubeService(si typedv1.ServiceInterface, labels labels.Set) (*apiv1.Service, error) {
	services, err := si.List(metav1.ListOptions{
		LabelSelector: labels.AsSelector().String(),
	})
	if err != nil {
		return nil, err
	}
	if l := len(services.Items); l == 0 {
		return nil, ErrServiceNotFound
	} else if l > 1 {
		return nil, ErrMultipleServicesFound
	}
	svc := services.Items[0]
	return &svc, nil
}

// routeToVersion queries kubernetets for service matching the provided
// logical name and version.
// If a valid service is found with the required annotations then a dentry is
// upserted and a dtab sync triggered to update the namerd dtab for the service
func routeToVersion(s ServiceSelectorDentryPutter, cs *kubernetes.Clientset, ln, ver string) (*pb.RouteToVersionResponse, error) {
	// Get service by it's logical name
	svc, err := s.GetByLogicalName(ln)
	if err != nil {
		return nil, err
	}
	// Get kubernetets service
	ksvc, err := getKubeService(
		cs.CoreV1().Services(svc.Namespace),
		labels.Set{
			"logicalName": svc.LogicalName,
			"version":     ver,
		})
	// Lookup the dtab annotation so we know what dtab the dentry is managed within
	dtab, ok := ksvc.Annotations[DtabAnnotation]
	if !ok {
		return nil, ErrMissingDtabAnnotation
	}
	// Lookup the port-name annotation so we know what port-name to store in the dentry
	portName, ok := ksvc.Annotations[PortNameAnnotation]
	if !ok {
		return nil, ErrMissingPortNameAnnotation
	}
	// Lookup the prefix annotation with a default prefix being the service logical name
	prefix := fmt.Sprintf("/%s", svc.LogicalName)
	if prefixAnnotation, ok := ksvc.Annotations[DentryPrefixAnnotation]; ok {
		prefix = prefixAnnotation
	}
	// Lookup priority annotation so we we can place the dentry in the appropriate place in teh dtab
	priorityStr, ok := ksvc.Annotations[DentryPriorityAnnotation]
	if !ok {
		return nil, ErrMissingDentryPriorityAnnotation
	}
	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		return nil, errors.New("could not convert prioroty to int")
	}
	// Put the dentry to storage
	_, err = s.PutDentry(&store.Dentry{
		Prefix:      prefix,
		Destination: fmt.Sprintf("/#/io.l5d.k8s/%s/%s/%s", ksvc.Namespace, portName, ksvc.Name),
		Dtab:        dtab,
		Priority:    int32(priority),
	})
	if err != nil {
		return nil, err
	}
	// Return response
	return &pb.RouteToVersionResponse{}, nil
}

// DelegationTables returns a list of delgation tables managed by needle
func (n *DentryService) DelegationTables(ctx context.Context, req *pb.DelegationTablesRequest) (*pb.DelegationTablesResponse, error) {
	tables, err := n.store.DelegationTables()
	if err != nil {
		return nil, err
	}
	var dtabs []*pb.DelegationTable
	for _, t := range tables {
		dtabs = append(dtabs, &pb.DelegationTable{
			Name: t,
		})
	}
	return &pb.DelegationTablesResponse{
		DelegationTables: dtabs,
	}, nil
}

// Dentries returns a list of dentries for a delegation table
func (n *DentryService) Dentries(ctx context.Context, req *pb.DentriesRequest) (*pb.DentriesResponse, error) {
	dCh, err := n.store.DentriesByDtab(ctx, req.GetDtab())
	if err != nil {
		return nil, err
	}
	var dentries []*pb.Dentry
	for dentry := range dCh {
		dentries = append(dentries, &pb.Dentry{
			Id:          dentry.Id.String(),
			CreateDate:  &timestamp.Timestamp{Seconds: dentry.CreateDate.Unix()},
			UpdateDate:  &timestamp.Timestamp{Seconds: dentry.UpdateDate.Unix()},
			Dtab:        dentry.Dtab,
			Prefix:      dentry.Prefix,
			Destination: dentry.Destination,
			Priority:    dentry.Priority,
		})
	}
	return &pb.DentriesResponse{
		Dentries: dentries,
	}, nil
}

func (ds *DentryService) RegisterWithServer(srv *grpc.Server) {
	pb.RegisterDentryServiceServer(srv, ds)
}

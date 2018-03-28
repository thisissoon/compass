package service

import (
	"context"

	pb "compass/pkg/proto/services"

	"github.com/mattes/migrate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// migrator interfaces
type (
	upgrader interface {
		Up() error
	}
	versioner interface {
		Version() (uint, bool, error)
	}
	migrator interface {
		upgrader
		versioner
	}
)

type DataMigrationService struct {
	migrator migrator
}

func NewDataMigrationService(m *migrate.Migrate) *DataMigrationService {
	return &DataMigrationService{
		migrator: m,
	}
}

func (dms *DataMigrationService) Upgrade(ctx context.Context, req *pb.DataMigrationUpgradeRequest) (*pb.DataMigrationUpgradeResponse, error) {
	if err := dms.migrator.Up(); err != nil {
		return nil, status.Errorf(codes.Internal, "data migration upgrade error: %s", err)
	}
	return &pb.DataMigrationUpgradeResponse{}, nil
}

func (dms *DataMigrationService) Version(ctx context.Context, req *pb.DataMigrationVersionRequest) (*pb.DataMigrationVersionResponse, error) {
	version, _, err := dms.migrator.Version()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get schema version: %s", err)
	}
	return &pb.DataMigrationVersionResponse{
		Version: uint32(version),
	}, nil
}

func (dms *DataMigrationService) RegisterWithServer(srv *grpc.Server) {
	pb.RegisterDataMigrationServiceServer(srv, dms)
}

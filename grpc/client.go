package grpc

import (
	"time"

	"google.golang.org/grpc"

	"compass/logger"

	needle "compass/proto/needle/v1"
)

func NewClient(addr string) (needle.NeedleServiceClient, error) {
	log := logger.New()
	log.Debug().Str("address", addr).Msg("connecting to gRPC server")
	cc, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*5),
	)
	if err != nil {
		return nil, err
	}
	return needle.NewNeedleServiceClient(cc), nil
}

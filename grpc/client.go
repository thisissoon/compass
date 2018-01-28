package grpc

import (
	"time"

	"google.golang.org/grpc"

	"compass/logger"

	needle "compass/proto/needle/v1"
)

func NewClient(addr string) (needle.NeedleServiceClient, bool) {
	log := logger.New()
	log.Debug().Str("address", addr).Msg("connecting to gRPC server")
	cc, err := grpc.Dial(
		addr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*5),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create gRPC connection")
		return nil, false
	}
	return needle.NewNeedleServiceClient(cc), true
}

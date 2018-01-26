package grpc

import (
	"time"

	"google.golang.org/grpc"

	"compass/logger"

	needle "compass/proto/needle/v1"
)

func NewClient(addr string) (needle.NeedleServiceClient, bool) {
	log := logger.New()
	cc, err := grpc.Dial(
		"localhost:5000",
		grpc.WithInsecure(),
		grpc.WithAuthority("compass"),
		grpc.WithTimeout(time.Second*5),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create gRPC connection")
		return nil, false
	}
	return needle.NewNeedleServiceClient(cc), true
}

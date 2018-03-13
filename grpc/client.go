package grpc

import (
	"time"

	"google.golang.org/grpc"

	needle "compass/proto/needle/v1"
)

func NewClient(addr string) (needle.NeedleServiceClient, error) {
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

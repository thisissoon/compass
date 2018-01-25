package grpc

import (
	"google.golang.org/grpc"

	needle "compass/proto/needle/v1"
)

func ClientConn(addr string) (*grpc.ClientConn, error) {
	do := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	return grpc.Dial(addr, do...)
}

func NewServiceManagerClient(cc *grpc.ClientConn) needle.NeedleServiceClient {
	return needle.NewNeedleServiceClient(cc)
}

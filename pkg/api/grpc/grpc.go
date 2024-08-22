package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mem-db/pkg/api"
	service "mem-db/pkg/service"
	"net"
)

type GRPCServer struct {
	server *grpc.Server
}

func NewServer(ctx context.Context) api.Server {
	return &GRPCServer{
		server: grpc.NewServer(),
	}
}

func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("Failed to listen on port 50051: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())
	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	return s.server.GracefulStop()
}

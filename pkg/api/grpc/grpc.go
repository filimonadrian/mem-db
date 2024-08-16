package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"mem-db/pkg/api"
	service "mem-db/pkg/service"
	"net"
)

type GRPCServer struct {
	grpcServer *grpc.Server
}

func NewServer(ctx context.Context, svc service.WordService) api.Server {
	return &GRPCServer{
		grpcServer: grpc.NewServer(),
	}
}

func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("Failed to listen on port 50051: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())
	return s.grpcServer.Serve(lis)
}

func (s *GRPCServer) Stop(ctx context.Context) error {
	return nil
}

package grpc

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"mem-db/pkg/api"
	"net"
)

type GRPCServer struct {
	grpcServer *grpc.Server
}

func NewServer() api.Server {
	return &GRPCServer{
		grpcServer: grpc.NewServer(),
	}
}

func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("Failed to listen on port 50051: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())
	return s.grpcServer.Serve(lis)
}

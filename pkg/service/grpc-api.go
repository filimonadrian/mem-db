package service

import (
	"context"
	config "mem-db/cmd/config"
)

type DBGrpcServer struct {
}

func NewDBGrpcServer(ctx context.Context, options *config.ServiceOptions, svc WordService) *DBGrpcServer {

	return nil
}

package service

import (
	"context"
	"fmt"
	config "mem-db/cmd/config"
	// api "mem-db/pkg/api"
	repo "mem-db/pkg/repository"
	// "time"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func InitService(ctx context.Context, config *config.Config) Service {
	db, wal, err := repo.InitDBSystem(ctx, &config.WALOptions)
	if err != nil {
		panic(fmt.Errorf("Error while trying to initialize DB system: %v", err.Error()))
	}

	wordService := NewWordService(ctx, db, wal)

	// var server api.Server
	var dbService Service

	// if config.ServiceOptions.ApiOptions.UseGRPC {
	// 	dbService = NewDBGrpcServer(ctx, config.ServiceOptions, wordService)
	// } else {
	// }
	dbService = NewDBHttpServer(ctx, &config.ServiceOptions, wordService)

	return dbService
}

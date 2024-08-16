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

	// handle gracefully shutdown
	// go func(ctx context.Context) {
	// 	select {
	// 	case <-ctx.Done():
	// 		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 4*time.Second)
	// 		defer shutdownRelease()
	// 		if err := dbService.Stop(shutdownCtx); err != nil {
	// 			// fmt.Printf("Error while stopping the server: %v\n", err.Error())
	// 			fmt.Println("Error while stopping the server: ", err.Error())

	// 		}
	// 	}
	// }(ctx)

	return dbService
}

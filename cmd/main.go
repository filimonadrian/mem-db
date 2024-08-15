package main

import (
	"context"
	"fmt"
	config "mem-db/cmd/config"
	api "mem-db/pkg/api"
	grpc "mem-db/pkg/api/grpc"
	httpserver "mem-db/pkg/api/http/server"
	repo "mem-db/pkg/repository"
	service "mem-db/pkg/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// func closeApp(ctx context.Context) {
// 	for {
// 		select <-
// 	}

// }

func main() {
	if len(os.Args) < 2 {
		panic("Config file path is missing!")
	}

	configFilePath := os.Args[1]

	config, err := config.ReadConfig(configFilePath)
	if err != nil {
		panic(fmt.Errorf("Error while trying to read service config: %v", err.Error()))
	}

	db, wal, err := repo.InitDBSystem(&config.WALOptions)
	if err != nil {
		panic(fmt.Errorf("Error while trying to initialize DB system: %v", err.Error()))
	}

	// check if app is being closed and close resources
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	svc := service.NewService(db, wal)

	var server api.Server

	if config.UseGRPC {
		server = grpc.NewServer()
	} else {
		server = httpserver.NewServer(config.Port, svc)
	}

	// handle gracefully shutdown
	go wal.KeepSyncing(ctx)
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 3*time.Second)
			defer shutdownRelease()
			if err := server.Stop(shutdownCtx); err != nil {
				// fmt.Printf("Error while stopping the server: %v\n", err.Error())
				fmt.Println("Error while stopping the server: ", err.Error())

			}
		}
	}(ctx)

	if err = server.Start(); err != nil {
		panic(fmt.Errorf("Error starting server: %v", err))
	}

}

package main

import (
	"context"
	"fmt"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
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

func main() {
	if len(os.Args) < 2 {
		panic("Config file path is missing!")
	}

	configFilePath := os.Args[1]

	config, err := config.ReadConfig(configFilePath)
	if err != nil {
		panic(fmt.Errorf("Error while trying to read service config: %v", err.Error()))
	}

	var logger log.Logger
	if config.LoggerOptions.Console {
		logger, err = log.NewConsoleLogger(&config.LoggerOptions)
	} else {
		logger, err = log.NewFileLogger(&config.LoggerOptions)
	}
	if err != nil {
		panic(fmt.Errorf("Error while trying to initialize logger: %v", err.Error()))
	}

	ctx := context.WithValue(context.Background(), log.LoggerKey, logger)

	db, wal, err := repo.InitDBSystem(ctx, &config.WALOptions)
	if err != nil {
		panic(fmt.Errorf("Error while trying to initialize DB system: %v", err.Error()))
	}

	// check if app is being closed and close resources
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	svc := service.NewService(ctx, db, wal)

	var server api.Server

	if config.ApiOptions.UseGRPC {
		server = grpc.NewServer(ctx, svc)
	} else {
		server = httpserver.NewServer(ctx, config.ApiOptions.Port, svc)
	}

	// handle gracefully shutdown
	go wal.KeepSyncing(stopCtx)
	go func(stopCtx context.Context) {
		select {
		case <-stopCtx.Done():
			shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 3*time.Second)
			defer shutdownRelease()
			if err := server.Stop(shutdownCtx); err != nil {
				// fmt.Printf("Error while stopping the server: %v\n", err.Error())
				fmt.Println("Error while stopping the server: ", err.Error())

			}
		}
	}(stopCtx)

	if err = server.Start(); err != nil {
		panic(fmt.Errorf("Error starting server: %v", err))
	}

}

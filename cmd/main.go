package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	node "mem-db/pkg/node"
	repo "mem-db/pkg/repository"
	service "mem-db/pkg/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Shutdown(ctx context.Context, wordService service.WordService, nodeService node.NodeService) error {
	select {
	case <-ctx.Done():
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 4*time.Second)
		defer shutdownRelease()
		if err := wordService.Stop(shutdownCtx); err != nil {
			return fmt.Errorf("Error while stopping DB service: %v", err.Error())
		}
		if err := nodeService.Stop(shutdownCtx); err != nil {
			return fmt.Errorf("Error while stopping Node Service: %v", err.Error())
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		panic("Config file path is missing!")
	}

	configFilePath := os.Args[1]

	config, err := config.ReadConfig(configFilePath)
	if err != nil {
		panic(fmt.Errorf("Error while trying to read application config: %v", err.Error()))
	}

	// check if app is being closed and close resources
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var logger log.Logger
	if config.LoggerOptions.Console {
		logger, err = log.NewConsoleLogger(&config.LoggerOptions)
	} else {
		logger, err = log.NewFileLogger(&config.LoggerOptions)
	}
	if err != nil {
		panic(fmt.Errorf("Error while trying to initialize logger: %v", err.Error()))
	}

	ctx = context.WithValue(ctx, log.LoggerKey, logger)
	nodeService := node.NewNodeService(ctx, &config.NodeOptions)

	dbService := repo.NewDatabase(ctx, config, nodeService.IsMaster())
	wordService := service.NewWordService(ctx, config, dbService)

	nodeService.SetRepo(dbService)
	nodeService.SetWS(wordService)

	var wg errgroup.Group

	// handle Database Start
	wg.Go(func() error {
		if err = wordService.Start(ctx); err != nil {
			return fmt.Errorf("Error starting DB Service: %v", err)
		}
		return nil
	})

	// handle NodeService Start
	wg.Go(func() error {
		if err = nodeService.Start(ctx); err != nil {
			return fmt.Errorf("Error starting Node Service: %v", err)
		}
		return nil
	})

	logger.Info("Successfully started node")

	//handle shutdown
	wg.Go(func() error {
		return Shutdown(ctx, wordService, nodeService)
	})

	if err := wg.Wait(); err != nil {
		logger.Error(err)
	}
}

package main

import (
	"fmt"
	config "mem-db/cmd/config"
	api "mem-db/pkg/api"
	grpc "mem-db/pkg/api/grpc"
	httpserver "mem-db/pkg/api/http/server"
	"os"
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

	var server api.Server

	if config.UseGRPC {
		server = grpc.NewServer()
	} else {
		server = httpserver.NewServer(config.Port)
	}

	if err = server.Start(); err != nil {
		panic(fmt.Errorf("Error starting server: %v\n", err))
	}

}

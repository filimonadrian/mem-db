package cmd

import (
	"encoding/json"
	"fmt"
	logger "mem-db/cmd/logger"
	api "mem-db/pkg/api"
	rep "mem-db/pkg/replication"
	repo "mem-db/pkg/repository"
	"os"
)

type Config struct {
	ApiOptions         api.ApiOptions         `json:apiOptions`
	Master             bool                   `json:"master"`
	WALOptions         repo.WALOptions        `json:"walOptions"`
	LoggerOptions      logger.LoggerOptions   `json:"loggerOptions"`
	ReplicationOptions rep.ReplicationOptions `json:"replicationOptions"`
}

func ReadConfig(filePath string) (*Config, error) {

	byteData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	var config Config

	err = json.Unmarshal(byteData, &config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v\n", err)
	}

	return &config, nil
}

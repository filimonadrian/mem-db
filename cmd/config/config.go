package config

import (
	"encoding/json"
	"fmt"
	logger "mem-db/cmd/logger"
	"os"
)

type ApiOptions struct {
	Port    int  `json:"port"`
	UseGRPC bool `json:"useGRPC"`
}

type ServiceOptions struct {
	ApiOptions *ApiOptions `json:"apiOptions"`
}

type ReplicationOptions struct {
	Master     bool `json:"master"`
	MaxWorkers int  `json:"maxWorkers"`
}

type WALOptions struct {
	WalFilePath  string `json:"walFilePath"`
	SyncTimer    int    `json:"syncTimer"`
	SyncMaxBytes int    `json:"syncMaxBytes"`
}

type Config struct {
	ServiceOptions     ServiceOptions       `json:"serviceOptions"`
	ReplicationOptions ReplicationOptions   `json:"replicationOptions"`
	WALOptions         WALOptions           `json:"walOptions"`
	LoggerOptions      logger.LoggerOptions `json:"loggerOptions"`
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

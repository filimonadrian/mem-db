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

type WALOptions struct {
	WalFilePath  string `json:"walFilePath"`
	SyncTimer    int    `json:"syncTimer"`
	Restore      bool   `json:"restore"`
	SyncMaxBytes int    `json:"syncMaxBytes"`
}

type SnapshotOptions struct {
	DirPath   string `json:"dirPath"`
	SyncTimer int    `json:"syncTimer"`
}

type NodeOptions struct {
	Name              string      `json:"name"`
	MasterID          string      `json:"masterID,omitempty"`
	HeartbeatInterval int         `json:"heartbeatInterval"`
	ApiOptions        *ApiOptions `json:"apiOptions"`
	LeaderElection    bool        `json:"leaderElection"`
}

type Config struct {
	ServiceOptions  ServiceOptions       `json:"serviceOptions"`
	SnapshotOptions SnapshotOptions      `json:"snapshotOptions"`
	WALOptions      WALOptions           `json:"walOptions"`
	NodeOptions     NodeOptions          `json:"nodeOptions"`
	LoggerOptions   logger.LoggerOptions `json:"loggerOptions"`
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

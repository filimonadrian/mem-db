package cmd

import (
	"encoding/json"
	"fmt"
	repo "mem-db/pkg/repository"
	"os"
)

type Config struct {
	Port       int             `json:"port"`
	UseGRPC    bool            `json:"useGRPC"`
	WALOptions repo.WALOptions `json:"walOptions"`
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

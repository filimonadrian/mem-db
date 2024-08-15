package api

import (
	"context"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type ApiOptions struct {
	Port    int  `json:"port"`
	UseGRPC bool `json:"useGRPC"`
}

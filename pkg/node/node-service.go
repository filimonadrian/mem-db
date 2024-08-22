package node

import (
	"context"
	config "mem-db/cmd/config"
	repo "mem-db/pkg/repository"
	service "mem-db/pkg/service"
)

type NodeService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsMaster() bool
	SetRepo(db repo.DBService)
	SetWS(ws service.WordService)
}

func NewNodeService(ctx context.Context, options *config.NodeOptions) NodeService {
	return NewNode(ctx, options)
}

package node

import (
	"context"
	config "mem-db/cmd/config"
)

type NodeService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func InitNodeService(ctx context.Context, options *config.NodeOptions) NodeService {
	node := NewNode(ctx, options)

	var nodeService NodeService
	if node.IsMaster() {
		nodeService := NewMasterHttpServer(ctx, options, node)
	} else {
		nodeService := NewWorkerHttpServer(ctx, options, node)
	}

	nodeService.Start()
}

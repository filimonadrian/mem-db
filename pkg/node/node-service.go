package node

import (
	"context"
	config "mem-db/cmd/config"
)

type NodeService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func NewNodeService(ctx context.Context, options *config.NodeOptions) NodeService {
	return NewNode(ctx, options)
}

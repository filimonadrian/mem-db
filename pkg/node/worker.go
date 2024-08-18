package node

import (
	"context"
	"encoding/json"
	"fmt"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	httpserver "mem-db/pkg/api/http/server"
	"net/http"
)

type WorkerHttpServer struct {
	server *httpserver.HTTPServer
	logger log.Logger
	// node   *Node
}

func NewWorkerHttpServer(ctx context.Context, options *config.NodeOptions, node *Node) *WorkerHttpServer {
	httpServer := &WorkerHttpServer{
		server: httpserver.NewServer(ctx, options.ApiOptions.Port),
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}

	httpServer.server.Router.AddRoute("POST", "/worker/workers-list", node.updateWorkersList)
	httpServer.server.Router.AddRoute("POST", "/worker/master-id", node.updateMasterID)
	httpServer.server.Router.AddRoute("POST", "/worker/heartbeat", node.heartbeat)

	return httpServer
}

func (s *WorkerHttpServer) Start() error {
	return s.server.Start()
}

func (s *WorkerHttpServer) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

func (n *Node) updateWorkersList(w http.ResponseWriter, r *http.Request) {
	var workersMap map[string]struct{}

	err := json.NewDecoder(r.Body).Decode(&workersMap)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body for workers list: %v", err), http.StatusBadRequest)
		return
	}

	// Update the workers list
	err = n.UpdateWorkersList(workersMap)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update workers list: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (n *Node) updateMasterID(w http.ResponseWriter, r *http.Request) {
	var masterID string

	err := json.NewDecoder(r.Body).Decode(&masterID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body for masterID: %v", err), http.StatusBadRequest)
		return
	}

	n.UpdateMasterID(masterID)
	w.WriteHeader(http.StatusOK)
}

func (n *Node) heartbeat(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

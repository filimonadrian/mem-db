package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	httpserver "mem-db/pkg/api/http/server"
	"net/http"
)

type WorkerHttpServer struct {
	server *httpserver.HTTPServer
	logger log.Logger
}

func NewWorkerHttpServer(ctx context.Context, options *config.NodeOptions, node *Node) *WorkerHttpServer {
	httpServer := &WorkerHttpServer{
		server: httpserver.NewServer(ctx, options.ApiOptions.Port),
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}

	httpServer.server.Router.AddRoute("POST", "/worker/workers-list", node.updateWorkersList)
	httpServer.server.Router.AddRoute("POST", "/worker/master-id", node.updateMasterID)
	httpServer.server.Router.AddRoute("POST", "/worker/master-database", node.loadMasterDatabase)
	httpServer.server.Router.AddRoute("GET", "/worker/heartbeat", node.heartbeat)

	return httpServer
}

func (s *WorkerHttpServer) Start() error {
	s.logger.Info("Starting Worker Node..")
	return s.server.Start()
}

func (s *WorkerHttpServer) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

func (n *Node) updateWorkersList(w http.ResponseWriter, r *http.Request) {
	var workersMap map[string]struct{}
	n.Logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	err := json.NewDecoder(r.Body).Decode(&workersMap)
	if err != nil {
		n.Logger.Error(fmt.Sprintf("Invalid request body for workers list: %v", err))
		http.Error(w, fmt.Sprintf("Invalid request body for workers list: %v", err), http.StatusBadRequest)
		return
	}

	// Update the workers list
	err = n.UpdateWorkersList(workersMap)
	if err != nil {
		n.Logger.Error(fmt.Sprintf("Failed to update workers list: %v", err))
		http.Error(w, fmt.Sprintf("Failed to update workers list: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (n *Node) updateMasterID(w http.ResponseWriter, r *http.Request) {
	var md *NodeDetails = &NodeDetails{}
	n.Logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	err := json.NewDecoder(r.Body).Decode(md)
	if err != nil {
		n.Logger.Error(fmt.Sprintf("Invalid request body for masterID: %v", err))
		http.Error(w, fmt.Sprintf("Invalid request body for masterID: %v", err), http.StatusBadRequest)
		return
	}

	n.UpdateMasterID(md.Name)
	w.WriteHeader(http.StatusOK)
}

func (n *Node) heartbeat(w http.ResponseWriter, r *http.Request) {
	n.Logger.Debug("Heartbeat")
	w.WriteHeader(http.StatusOK)
}

func (n *Node) loadMasterDatabase(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	var err error
	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(r.Body)
	if err != nil {
		n.Logger.Error(fmt.Sprintf("Error reading request body: %v", err))
		http.Error(w, fmt.Sprintf("Invalid body %v", err), http.StatusBadRequest)
		return
	}
	n.Logger.Info("Received encoded database from master")

	err = n.LoadDatastore(bodyBytes)
	if err != nil {
		n.Logger.Error(fmt.Sprintf("Error loading datastore: %v", err))
		http.Error(w, fmt.Sprintf("Error loading datastore %v", err), http.StatusInternalServerError)
		return
	}
	n.Logger.Info("Loaded database received from master")
	w.WriteHeader(http.StatusOK)
}

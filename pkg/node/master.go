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

type MasterHttpServer struct {
	server *httpserver.HTTPServer
	logger log.Logger
}

type WorkerDetails struct {
	Name string `json:"name"`
}

type Response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message,omitempty"`
}

func NewMasterHttpServer(ctx context.Context, options *config.NodeOptions, node *Node) *MasterHttpServer {
	httpServer := &MasterHttpServer{
		server: httpserver.NewServer(ctx, options.ApiOptions.Port),
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}
	httpServer.server.Router.AddRoute("POST", "/master/register", node.registerWorker)
	httpServer.server.Router.AddRoute("POST", "/master/replicate", node.replicate)

	return httpServer
}

func (s *MasterHttpServer) Start() error {
	return s.server.Start()
}

func (s *MasterHttpServer) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

func (n *Node) registerWorker(w http.ResponseWriter, r *http.Request) {
	var wd *WorkerDetails = &WorkerDetails{}

	err := json.NewDecoder(r.Body).Decode(wd)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error()})
		return
	}

	n.RegisterWorker(wd.Name)
	// send the db snapshot to the Worker
	// for now, send just the location of the file

	// broadcast the list of workers because it's changed
	err = n.BroadcastWorkersList()
	if err != nil {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error()})
		return
	}

	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
	})
}

// used to forward POST requests in database to the workers' client api
func (n *Node) replicate(w http.ResponseWriter, r *http.Request) {

	err := n.ForwardToWorkers(r)
	if err != nil {
		n.Logger.Error(fmt.Sprintf("Cannot replicate the request to the workers: %v", err.Error()))
		json.NewEncoder(w).Encode(&Response{
			Status:     "Error",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
	}

	n.Logger.Debug("Replicated request to the workers")
	w.WriteHeader(http.StatusOK)
}

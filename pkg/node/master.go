package node

import (
	// "bytes"
	"context"
	"encoding/json"
	"fmt"
	// "io"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	httpserver "mem-db/pkg/api/http/server"
	"net/http"
)

type MasterHttpServer struct {
	server *httpserver.HTTPServer
	logger log.Logger
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
	s.logger.Info("Starting Master Node..")
	return s.server.Start()
}

func (s *MasterHttpServer) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

func (n *Node) registerWorker(w http.ResponseWriter, r *http.Request) {
	var wd *NodeDetails = &NodeDetails{}
	n.Logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	err := json.NewDecoder(r.Body).Decode(wd)
	if err != nil {
		n.Logger.Error("Cannot decode workerDetails: ", err.Error())
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
		n.Logger.Error("Cannot broadcast the list of workers: ", err.Error())
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
	n.Logger.Info("Workers list was successfully broadcasted")
}

// used to forward POST requests in database to the workers' client api
func (n *Node) replicate(w http.ResponseWriter, r *http.Request) {
	n.Logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	var err error
	err = n.ForwardToWorkers(r)

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

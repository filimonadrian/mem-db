package node

import (
	// "bytes"
	"context"
	"encoding/json"
	"fmt"
	// "io"
	"errors"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	httpclient "mem-db/pkg/api/http/client"
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
	// httpServer.server.Router.AddRoute("POST", "/master/replicate", node.replicate)

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
	var err error

	n.Logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	err = json.NewDecoder(r.Body).Decode(wd)
	if err != nil {
		n.Logger.Error("Cannot decode workerDetails: ", err.Error())
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error()})
		return
	}

	n.RegisterWorker(wd.Name)
	// send post request to worker with encoded db
	workerUrl := httpclient.GetURL(wd.Name, n.Port, "/worker/master-database")
	encodedData, err := n.GetEncodedDatastore()
	if err != nil {
		n.Logger.Error("Cannot encode database: ", err)
		json.NewEncoder(w).Encode(&Response{
			Status:     "Error",
			StatusCode: http.StatusInternalServerError,
			Message:    "Errors when encoding database",
		})

		return
	}

	err = httpclient.SendPostRequest(workerUrl, encodedData)
	if err != nil {
		n.Logger.Error("Cannot send encoded database: ", err)
	}
	n.Logger.Info("Sent encoded database to the new worker ", wd.Name)

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

func (n *Node) replicateToWorkers(ctx context.Context, isGRPC bool) {
	for {
		select {
		case data := <-n.forwardingCh:
			if isGRPC {
				err := n.ForwardToWorkersGRPC(data)
				if err != nil {
					n.Logger.Error("Cannot forward data to workers: ", err)
				}
			} else {
				err := n.ForwardToWorkersHTTP(data)
				if err != nil {
					n.Logger.Error("Cannot forward data to workers: ", err)
				}
			}
		case <-ctx.Done():
			close(n.forwardingCh)
			return
		}
	}
}

func (n *Node) ForwardToWorkersGRPC(data []byte) error {
	return nil
}

// forward requests with words to the workers
func (n *Node) ForwardToWorkersHTTP(data []byte) error {
	n.Logger.Debug("Started forwarding the request to the workers: ", n.Workers)

	var errs error
	for workerName, _ := range n.Workers {
		forwardURL := httpclient.GetURL(workerName, 8080, "/words/register")
		n.Logger.Debug("Forwarding the request to  ", forwardURL)

		err := httpclient.SendPostRequest(forwardURL, data)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to forward request to worker %s: %v", forwardURL, err))
		}
	}

	if errs != nil {
		return errs
	}

	return nil
}

// call from master to add worker to the workers list
func (n *Node) RegisterWorker(workerName string) {
	n.Logger.Info(fmt.Sprintf("Registered node %s as worker", workerName))
	if len(n.Workers) >= 0 {
		n.SetForwarding()
	}
	n.Workers[workerName] = struct{}{}
}

func (n *Node) DeleteWorker(workerName string) {
	n.Logger.Warn(fmt.Sprintf("Deleted worker %s from list", workerName))
	delete(n.Workers, workerName)

	if len(n.Workers) == 0 {
		n.UnsetForwarding()
	}

	n.Logger.Warn("Active Workers: ", n.Workers)
}

// send the list of workers to every other worker
// workers should delete themselves from the list
func (n *Node) BroadcastWorkersList() error {

	payload, err := json.Marshal(n.Workers)
	if err != nil {
		return fmt.Errorf("Error marshaling workers list: %v\n", err)
	}
	n.Logger.Info("Broadcasting workers list..")
	err = n.broadcast("/worker/workers-list", payload)
	if err != nil {
		return fmt.Errorf("Errors broadcasting the workers list: %v\n", err)
	}

	return nil
}

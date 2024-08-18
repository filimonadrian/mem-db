package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	api "mem-db/pkg/api"
	httpclient "mem-db/pkg/api/http/client"
	"net/http"
	"time"
)

type Node struct {
	Name              string
	MasterID          string
	PartitionMasters  []string
	Workers           map[string]struct{}
	Port              int
	HeartbeatInterval int
	Logger            log.Logger
	Server            api.Server
}

func NewNode(ctx context.Context, options *config.NodeOptions) *Node {
	node := &Node{
		Name:              options.Name,
		MasterID:          options.MasterID,
		Workers:           make(map[string]struct{}),
		PartitionMasters:  make([]string, 0),
		Logger:            ctx.Value(log.LoggerKey).(log.Logger),
		Port:              options.ApiOptions.Port,
		HeartbeatInterval: options.HeartbeatInterval,
	}

	if node.IsMaster() {
		node.Server = NewMasterHttpServer(ctx, options, node)
	} else {
		node.Server = NewWorkerHttpServer(ctx, options, node)
	}

	return node
}

func (n *Node) Start(ctx context.Context) error {
	if n.IsMaster() {
		go n.StartHeartbeatCheck(ctx, time.Duration(n.HeartbeatInterval)*time.Second)
	}
	return n.Server.Start()
}

func (n *Node) Stop(ctx context.Context) error {
	return n.Server.Stop(ctx)
}

func (n *Node) IsMaster() bool {
	return n.MasterID == ""
}

func (n *Node) RegisterWorker(workerName string) {
	n.Workers[workerName] = struct{}{}
}

// this will delete a worker which did not repond to my heartbeat
func (n *Node) DeleteWorker(workerName string) {
	delete(n.Workers, workerName)
}

// send the list of workers to every other worker
// workers should delete themselfs from the list
func (n *Node) BroadcastWorkersList() error {

	payload, err := json.Marshal(n.Workers)
	if err != nil {
		return fmt.Errorf("Error marshaling workers list: %v\n", err)
	}
	err = n.broadcast("/worker/workers-list", payload)
	if err != nil {
		return fmt.Errorf("Errors broadcasting the workers list: %v\n", err)
	}

	return nil
}

// special request when I'm elected as leader
func (n *Node) BroadcastMasterID() error {
	payload, err := json.Marshal(n.Name)
	if err != nil {
		return fmt.Errorf("Error marshaling Master's Name: %v\n", err)
	}
	err = n.broadcast("/worker/master-id", payload)
	if err != nil {
		return fmt.Errorf("Errors broadcasting the master's Name: %v\n", err)
	}

	return nil
}

func (n *Node) broadcast(endpoint string, payload []byte) error {

	var allErrs error
	for address, _ := range n.Workers {
		url := httpclient.GetURL(address, n.Port, endpoint)
		err := httpclient.SendPostRequest(url, payload)
		if err != nil {
			n.Logger.Error(err.Error())
			allErrs = errors.Join(allErrs, err)
		}
	}
	if allErrs != nil {
		return allErrs
	}

	return nil
}

// forward requests with words to the workers
func (n *Node) ForwardToWorkers(originalRequest *http.Request) error {

	for workerName, _ := range n.Workers {
		forwardURL := httpclient.GetURL(workerName, n.Port, originalRequest.RequestURI)
		err := httpclient.ForwardRequest(originalRequest, forwardURL)
		if err != nil {
			return fmt.Errorf("Failed to forward request to worker %s: %v", forwardURL, err)
		}
	}

	return nil
}

func (n *Node) UpdateMasterID(masterID string) {
	n.MasterID = masterID
}

func (n *Node) UpdateWorkersList(workers map[string]struct{}) error {

	// delete myselft from workers list
	delete(workers, n.Name)
	n.Workers = workers

	return nil
}

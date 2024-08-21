package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
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

type NodeDetails struct {
	Name string `json:"name"`
}

type Response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message,omitempty"`
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
	var wg errgroup.Group

	wg.Go(func() error {
		if err := n.Server.Start(); err != nil {
			return fmt.Errorf("Error starting node-service server: %v", err)
		}
		return nil
	})

	// wait for the api server to start
	time.Sleep(2 * time.Second)

	if n.IsMaster() {
		n.Logger.Info("Starting Hearbeat process..")
		err := n.activateForwarding()
		if err != nil {
			n.Logger.Error(err.Error())
		}
		go n.StartHeartbeatCheck(ctx, time.Duration(n.HeartbeatInterval)*time.Second)
	} else {
		err := n.SendRegistrationReq()
		if err != nil {
			return fmt.Errorf("Error registering to master node: %v", err)
		}
	}

	if err := wg.Wait(); err != nil {
		n.Logger.Error(err)
		return err
	}

	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	return n.Server.Stop(ctx)
}

func (n *Node) IsMaster() bool {
	return n.MasterID == ""
}

func (n *Node) runAsMaster() {

}

func Retry(ctx context.Context, f func() error, retryAttempts int, interval time.Duration) error {
	var err error

	for i := 0; i < retryAttempts; i++ {
		err := f()
		if err == nil {
			return nil
		}

		select {
		case <-time.After(interval):

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("All retries failed: %v", err)
}

func (n *Node) activateForwarding() error {
	dbServiceURL := httpclient.GetURL("localhost", 8080, "/words/forward")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	forwardingFunc := func() error {
		resp, err := httpclient.SendGetRequest(ctx, dbServiceURL, nil)
		if err != nil {
			return fmt.Errorf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
		return nil
	}

	if err := Retry(ctx, forwardingFunc, 3, 1*time.Second); err != nil {
		return fmt.Errorf("Cannot start forwarding to workers: %v", err)
	}

	return nil
}

// call from master to add worker to the workers list
func (n *Node) RegisterWorker(workerName string) {
	n.Logger.Info(fmt.Sprintf("Registered node %s as worker", workerName))
	n.Workers[workerName] = struct{}{}
}

func (n *Node) DeleteWorker(workerName string) {
	n.Logger.Warn(fmt.Sprintf("Deleted worker %s from list", workerName))
	delete(n.Workers, workerName)
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

// special request when I'm elected as leader
func (n *Node) BroadcastMasterID() error {
	payload, err := json.Marshal(NodeDetails{Name: n.Name})
	if err != nil {
		return fmt.Errorf("Error marshaling Master's Name: %v\n", err)
	}
	n.Logger.Info("Sending master-id to workers..", string(payload))
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

	n.Logger.Debug("Started forwarding the request to the workers: ", n.Workers)
	var errs error
	for workerName, _ := range n.Workers {
		forwardURL := httpclient.GetURL(workerName, 8080, "/words/register")
		n.Logger.Debug("Forwarding the request to  ", forwardURL)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		err := httpclient.ForwardRequest(ctx, originalRequest, forwardURL)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to forward request to worker %s: %v", forwardURL, err))
		}
	}

	if errs != nil {
		return errs
	}
	return nil
}

func (n *Node) UpdateMasterID(masterID string) {
	n.Logger.Info("Updated MasterID to ", masterID)
	n.MasterID = masterID
}

func (n *Node) UpdateWorkersList(workers map[string]struct{}) error {

	// delete myselft from workers list
	delete(workers, n.Name)
	n.Workers = workers
	n.Logger.Info("Updated workers list: ", workers)

	return nil
}

// worker request for registration to master
func (n *Node) SendRegistrationReq() error {
	n.Logger.Info(fmt.Sprintf("Registering worker %s to master %s", n.Name, n.MasterID))

	// send POST request with
	url := httpclient.GetURL(n.MasterID, n.Port, "/master/register")

	payload, err := json.Marshal(NodeDetails{Name: n.Name})
	if err != nil {
		return fmt.Errorf("Error marshaling Master's Name: %v\n", err)
	}

	err = httpclient.SendPostRequest(url, payload)
	if err != nil {
		return fmt.Errorf("Cannot register the worker %s to the master %s: %v", n.Name, n.MasterID, err)
	}

	n.Logger.Info("Successfully registered to the master ", n.MasterID)
	return nil
}

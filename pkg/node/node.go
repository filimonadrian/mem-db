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
	repo "mem-db/pkg/repository"
	service "mem-db/pkg/service"
	// util "mem-db/pkg/util"
	"time"
)

type Node struct {
	Name              string
	MasterID          string
	PartitionMasters  map[string]struct{}
	Workers           map[string]struct{}
	Port              int
	HeartbeatInterval int
	Logger            log.Logger
	forwardingCh      chan []byte
	Server            api.Server
	db                repo.DBService
	ws                service.WordService
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
		PartitionMasters:  make(map[string]struct{}),
		Logger:            ctx.Value(log.LoggerKey).(log.Logger),
		Port:              options.ApiOptions.Port,
		HeartbeatInterval: options.HeartbeatInterval,
		forwardingCh:      make(chan []byte),
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
		n.SetForwardingCh()
		n.Logger.Info("Starting Hearbeat process..")
		go n.StartHeartbeatCheck(ctx, time.Duration(n.HeartbeatInterval)*time.Second)
		go n.replicateToWorkers(ctx, false)
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
	close(n.forwardingCh)
	return n.Server.Stop(ctx)
}

func (n *Node) SetRepo(db repo.DBService) {
	n.db = db
}

func (n *Node) SetWS(ws service.WordService) {
	n.ws = ws
}

func (n *Node) IsMaster() bool {
	return n.MasterID == ""
}

func (n *Node) runAsMaster() {

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

func (n *Node) LoadDatastore(encodedData []byte) error {
	return n.db.LoadDatastore(encodedData)
}

func (n *Node) GetEncodedDatastore() ([]byte, error) {
	return n.db.EncodeDatastore()
}

// methods for interacting with words service
func (n *Node) SetForwarding() {
	n.ws.SetForwarding()
}
func (n *Node) SetForwardingCh() {
	n.ws.SetForwardingCh(n.forwardingCh)
}

func (n *Node) UnsetForwarding() {
	n.ws.UnsetForwarding()
}

package node

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	leaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	config "mem-db/cmd/config"
	"os"
	"time"
)

func getClusterClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return clientset, nil
}

func (n *Node) leaderElection(ctx context.Context, options *config.NodeOptions) error {

	clientset, err := getClusterClientSet()
	if err != nil {
		return fmt.Errorf("Error getting clientset: %v\n", err)
	}

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "leader-election",
			Namespace: "default",
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: n.Name,
		},
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				n.runAsMaster(ctx, options)
			},
			OnStoppedLeading: func() {
				// This node is no longer the leader
				n.Logger.Warn("Lost leadership, exiting..")
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == n.Name {
					n.Logger.Info("Node elected as new leader: ", identity)
					// err := n.BroadcastMasterID()
				} else {
					n.Logger.Info("New leader elected", identity)
					n.UpdateMasterID(identity)
				}
			},
		},
		ReleaseOnCancel: true,
	})

	return nil
}

func (n *Node) runAsMaster(ctx context.Context, options *config.NodeOptions) {
	err := n.Server.Stop(ctx)
	if err != nil {
		n.Logger.Error("Server does not stop gracefully: ", err.Error())
	}

	n.Server = NewMasterHttpServer(ctx, options, n)
	n.MasterID = ""
	n.Start(ctx)

}

func runNewLeader() {}

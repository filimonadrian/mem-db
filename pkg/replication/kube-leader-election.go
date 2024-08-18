package replication

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	leaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
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

func leaderElection(ctx context.Context) error {

	// logger :=
	clientset, err := getClusterClientSet()
	if err != nil {
		return fmt.Errorf("Error getting clientset: %v\n", err)
	}

	lock := &resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{
			Name:      "leader-election",
			Namespace: "default",
		},
		Client: clientset.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: "node-1",
		},
	}

	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// This node is the leader now
				runAsMaster()
			},
			OnStoppedLeading: func() {
				// This node is no longer the leader
				fmt.Println("Lost leadership, exiting")
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				// notified when new leader elected
				if identity == id {
					// we have the lock
					runNewLeader()
				}
				fmt.Println("New leader elected", identity)
			},
		},
	})

}

func runAsMaster() {

}

func runNewLeader() {}

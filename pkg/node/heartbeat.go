package node

import (
	"context"
	"fmt"
	httpclient "mem-db/pkg/api/http/client"
	"net/http"
	"sync"
	"time"
)

func (n *Node) sendHeartbeat(ctx context.Context, workerName string) bool {
	workerURL := httpclient.GetURL(workerName, n.Port, "/worker/heartbeat")
	resp, err := httpclient.SendGetRequest(ctx, workerURL, nil)

	if err != nil || resp.StatusCode != http.StatusOK {
		n.Logger.Warn(fmt.Sprintf("Worker %s not responding", workerURL))
		return false
	}

	return true
}

func (n *Node) CheckAndRemoveDeadWorkers() {
	var wg sync.WaitGroup
	deadWorkers := make(chan string, len(n.Workers))

	for worker := range n.Workers {
		wg.Add(1)
		go func(worker string) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if !n.sendHeartbeat(ctx, worker) {
				deadWorkers <- worker
			}
		}(worker)
	}

	// Close deadWorkers channel once all goroutines are done
	go func() {
		wg.Wait()
		close(deadWorkers)
	}()

	// Remove dead workers
	for worker := range deadWorkers {
		n.Logger.Warn("Removing worker due to heartbeat failure: ", worker)
		n.DeleteWorker(worker)
	}
}

// Start the heartbeat check
func (n *Node) StartHeartbeatCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.CheckAndRemoveDeadWorkers()
		case <-ctx.Done():
			n.Logger.Debug("Heartbeat process stopped")
			return
		}
	}
}

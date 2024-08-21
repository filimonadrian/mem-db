package node

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	log "mem-db/cmd/logger"
	httpclient "mem-db/pkg/api/http/client"
)

func TestBroadcastWorkersList(t *testing.T) {
	// Step 1: Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request URL
		if r.URL.Path != "/worker/workers-list" {
			t.Errorf("Expected URL path /worker/workers-list, got %s", r.URL.Path)
		}

		// Verify the request payload
		var receivedPayload map[string]struct{}
		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		defer r.Body.Close()

		expectedPayload := map[string]struct{}{"worker1": {}, "worker2": {}}
		if len(receivedPayload) != len(expectedPayload) {
			t.Errorf("Expected payload length %d, got %d", len(expectedPayload), len(receivedPayload))
		}

		for worker := range expectedPayload {
			if _, ok := receivedPayload[worker]; !ok {
				t.Errorf("Expected worker %s in payload, but it was missing", worker)
			}
		}

		// Respond with 200 OK
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Step 2: Initialize the Node with mock workers
	options := &log.LoggerOptions{
		LogLevel:    "info",
		LogFilepath: "mem-db/data/mem-db.log",
		Console:     true,
	}
	logger, _ := log.NewConsoleLogger(options)

	node := &Node{
		Name:    "master",
		Workers: map[string]struct{}{"worker1": {}, "worker2": {}},
		Port:    8081,
		Logger:  logger,
	}

	// Override the GetURL function to use the mock server's URL
	httpclient.GetURL = func(address string, port int, endpoint string) string {
		return mockServer.URL + endpoint
	}

	// Step 3: Call the method and verify it doesn't return an error
	err := node.BroadcastWorkersList()
	if err != nil {
		t.Fatalf("BroadcastWorkersList() returned an error: %v", err)
	}
}

// Test BroadcastMasterID
func TestBroadcastMasterID(t *testing.T) {
	// Step 1: Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request URL
		if r.URL.Path != "/worker/master-id" {
			t.Errorf("Expected URL path /worker/master-id, got %s", r.URL.Path)
		}

		// Verify the request payload
		var receivedPayload string
		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		defer r.Body.Close()

		expectedPayload := "master"
		if receivedPayload != expectedPayload {
			t.Errorf("Expected payload %s, got %s", expectedPayload, receivedPayload)
		}

		// Respond with 200 OK
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Step 2: Initialize the Node with mock workers
	options := &log.LoggerOptions{
		LogLevel:    "info",
		LogFilepath: "mem-db/data/mem-db.log",
		Console:     true,
	}
	logger, _ := log.NewConsoleLogger(options)

	node := &Node{
		Name:    "master",
		Workers: map[string]struct{}{"worker1": {}, "worker2": {}},
		Port:    8081,
		Logger:  logger,
	}

	// Override the GetURL function to use the mock server's URL
	httpclient.GetURL = func(address string, port int, endpoint string) string {
		return mockServer.URL + endpoint
	}

	// Step 3: Call the method and verify it doesn't return an error
	err := node.BroadcastMasterID()
	if err != nil {
		t.Fatalf("BroadcastMasterID() returned an error: %v", err)
	}
}

// Test broadcast
func TestBroadcast(t *testing.T) {
	// Step 1: Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with 200 OK
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Step 2: Initialize the Node with mock workers
	options := &log.LoggerOptions{
		LogLevel:    "info",
		LogFilepath: "mem-db/data/mem-db.log",
		Console:     true,
	}
	logger, _ := log.NewConsoleLogger(options)

	node := &Node{
		Name:    "master",
		Workers: map[string]struct{}{"worker1": {}, "worker2": {}},
		Port:    8081,
		Logger:  logger,
	}

	// Override the GetURL function to use the mock server's URL
	httpclient.GetURL = func(address string, port int, endpoint string) string {
		return mockServer.URL + endpoint
	}

	// Step 3: Call the method and verify it doesn't return an error
	payload := []byte(`{"key":"value"}`)
	err := node.broadcast("/test-endpoint", payload)
	if err != nil {
		t.Fatalf("broadcast() returned an error: %v", err)
	}
}

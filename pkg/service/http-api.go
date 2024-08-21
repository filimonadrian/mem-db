package service

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	"time"
	// api "mem-db/pkg/api"
	"bytes"
	"io"
	httpclient "mem-db/pkg/api/http/client"
	httpserver "mem-db/pkg/api/http/server"
	"net/http"
)

type DBHttpServer struct {
	server     *httpserver.HTTPServer
	logger     log.Logger
	wordSvc    WordService
	forwarding bool
}

type TextInput struct {
	Text string `json:"text"`
}

type Response struct {
	Status     string         `json:"status"`
	StatusCode int            `json:"statusCode"`
	Data       []WordResponse `json:"data,omitempty"`
	Message    string         `json:"message,omitempty"`
}

func NewDBHttpServer(ctx context.Context, options *config.ServiceOptions, svc WordService) Service { // api.Server {

	dbHttpServer := &DBHttpServer{
		server:     httpserver.NewServer(ctx, options.ApiOptions.Port),
		wordSvc:    svc,
		logger:     ctx.Value(log.LoggerKey).(log.Logger),
		forwarding: false,
	}

	dbHttpServer.server.Router.AddRoute("GET", "/words/occurences", dbHttpServer.getWordOccurences)
	dbHttpServer.server.Router.AddRoute("POST", "/words/register", dbHttpServer.registerWords)
	// this endpoint is especially for syncing the servers
	// if i get a request, I will start to forward the request to the node-service
	dbHttpServer.server.Router.AddRoute("GET", "/words/forward", dbHttpServer.startForwarding)

	return dbHttpServer
}

func (s *DBHttpServer) Start(ctx context.Context) error {
	return s.server.Start()
}

func (s *DBHttpServer) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

// GET /words/occurences?terms=apple,banana,orange
func (s *DBHttpServer) getWordOccurences(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	terms := query["terms"]

	s.logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	if len(terms) == 0 {
		s.logger.Error("No words provided into request")
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    "No words provided into request"})
		return
	}

	results := s.wordSvc.GetOccurences(terms[0])

	s.logger.Debug("Results of the request: ", results)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Data:       results})

}

func (s *DBHttpServer) registerWords(w http.ResponseWriter, r *http.Request) {

	s.logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL))

	var bodyBytes []byte
	if r.Body == nil {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    "Body is empty"})
		return
	}

	var err error
	bodyBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	// Reset the body for the original request so it can be decoded
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var textInput *TextInput = &TextInput{}

	err = json.NewDecoder(r.Body).Decode(textInput)
	if err != nil {
		s.logger.Error("Cannot decode incoming request: ", err)
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error()})
		return
	}

	s.logger.Debug("Data from request: ", textInput)
	if len(textInput.Text) == 0 {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    "Text field is empty"})
		return
	}

	var wg errgroup.Group

	// Reset the body for the original request so it can be decoded
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if s.forwarding {
		wg.Go(func() error {
			s.logger.Info("Sending the INSERT request to the nodes service..")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err := httpclient.ForwardRequest(ctx, r, "http://localhost:8081/master/replicate")
			if err != nil {
				return fmt.Errorf("Cannot forward the request to the /master/replicate endpoint ", err)
			}

			return nil
		})
	}

	s.wordSvc.RegisterWords(textInput.Text)

	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Message:    "Text processed successfully"})

	if err = wg.Wait(); err != nil {
		s.logger.Error(err)
	}
}

func (s *DBHttpServer) startForwarding(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Received forwarding flag. All post request will be forwarded to the workers")
	s.forwarding = true
	w.WriteHeader(http.StatusOK)
}

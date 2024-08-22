package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	api "mem-db/pkg/api"
	// httpclient "mem-db/pkg/api/http/client"
	// "time"
	httpserver "mem-db/pkg/api/http/server"
	"net/http"
)

type DBHttpServer struct {
	server *httpserver.HTTPServer
	logger log.Logger
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

func NewDBHttpServer(ctx context.Context, options *config.ServiceOptions, ws *wordService) api.Server {

	dbHttpServer := &DBHttpServer{
		server: httpserver.NewServer(ctx, options.ApiOptions.Port),
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}

	dbHttpServer.server.Router.AddRoute("GET", "/words/occurences", ws.getWordOccurences)
	dbHttpServer.server.Router.AddRoute("POST", "/words/register", ws.registerWords)

	return dbHttpServer
}

func (s *DBHttpServer) Start() error {
	s.logger.Info("Starting server for WordService")
	return s.server.Start()
}

func (s *DBHttpServer) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

// GET /words/occurences?terms=apple,banana,orange
func (s *wordService) getWordOccurences(w http.ResponseWriter, r *http.Request) {
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

	results := s.GetOccurences(terms[0])

	s.logger.Debug("Results of the request: ", results)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Data:       results})

}

func (s *wordService) registerWords(w http.ResponseWriter, r *http.Request) {

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

	var textInput TextInput

	err = json.Unmarshal(bodyBytes, &textInput)
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

	if s.forwarding {
		s.forwardingCh <- bodyBytes
	}

	s.RegisterWords(textInput.Text)

	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Message:    "Text processed successfully"})
}

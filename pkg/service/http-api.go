package service

import (
	"context"
	"encoding/json"
	"fmt"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	// api "mem-db/pkg/api"
	httpserver "mem-db/pkg/api/http/server"
	"net/http"
)

type DBHttpServer struct {
	server  *httpserver.HTTPServer
	logger  log.Logger
	wordSvc WordService
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
		server:  httpserver.NewServer(ctx, options.ApiOptions.Port),
		wordSvc: svc,
		logger:  ctx.Value(log.LoggerKey).(log.Logger),
	}

	dbHttpServer.server.Router.AddRoute("GET", "/words/occurences", dbHttpServer.getWordOccurences)
	dbHttpServer.server.Router.AddRoute("POST", "/words/register", dbHttpServer.registerWords)

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
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    "No words provided into request"})
		return
	}

	results := s.wordSvc.GetOccurences(terms[0])

	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(results)
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Data:       results})

}

func (s *DBHttpServer) registerWords(w http.ResponseWriter, r *http.Request) {
	var textInput *TextInput = &TextInput{}

	err := json.NewDecoder(r.Body).Decode(textInput)
	if err != nil {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    err.Error()})
		return

	}

	if len(textInput.Text) == 0 {
		json.NewEncoder(w).Encode(&Response{
			Status:     "Bad Request",
			StatusCode: http.StatusBadRequest,
			Message:    "Text field is empty"})
		return
	}

	s.wordSvc.RegisterWords(textInput.Text)

	// json.NewEncoder(w).Encode(fmt.Sprintf("Text processed successfully"))
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Message:    "Text processed successfully"})
}

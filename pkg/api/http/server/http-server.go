package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "mem-db/cmd/logger"
	api "mem-db/pkg/api"
	router "mem-db/pkg/api/http/router"
	service "mem-db/pkg/service"
	"net/http"
)

type TextInput struct {
	Text string `json:"text"`
}

type Response struct {
	Status     string                 `json:"status"`
	StatusCode int                    `json:"statusCode"`
	Data       []service.WordResponse `json:"data,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

type HTTPServer struct {
	server  *http.Server
	service service.WordService
	logger  log.Logger
}

func NewServer(ctx context.Context, port int, svc service.WordService) api.Server {
	r := router.NewRouter()

	server := &HTTPServer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
		service: svc,
		logger:  ctx.Value(log.LoggerKey).(log.Logger),
	}

	r.AddRoute("GET", "/words/occurences", server.getWordOccurences)
	r.AddRoute("POST", "/words/register", server.registerWords)

	return server
}

func (s *HTTPServer) Start() error {

	s.logger.Info("Http server listening on port ", s.server.Addr)
	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %v", err)
	}
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {

	s.logger.Info("Shutting down http server on port ", s.server.Addr)
	return s.server.Shutdown(ctx)
}

// GET /words/occurences?terms=apple,banana,orange
func (s *HTTPServer) getWordOccurences(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())
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

	results := s.service.GetOccurences(terms[0])

	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(results)
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Data:       results})

}

func (s *HTTPServer) registerWords(w http.ResponseWriter, r *http.Request) {
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

	s.service.RegisterWords(textInput.Text)

	// json.NewEncoder(w).Encode(fmt.Sprintf("Text processed successfully"))
	json.NewEncoder(w).Encode(&Response{
		Status:     "Success",
		StatusCode: http.StatusOK,
		Message:    "Text processed successfully"})
}

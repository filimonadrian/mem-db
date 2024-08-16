package server

import (
	"context"
	"errors"
	"fmt"
	log "mem-db/cmd/logger"
	// api "mem-db/pkg/api"
	router "mem-db/pkg/api/http/router"
	"net/http"
)

type HTTPServer struct {
	Server *http.Server
	Router *router.Router
	logger log.Logger
}

func NewServer(ctx context.Context, port int) *HTTPServer { // api.Server {
	r := router.NewRouter()

	server := &HTTPServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: r,
		},
		Router: r,
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}

	return server
}

func (s *HTTPServer) Start() error {

	s.logger.Info("Http server listening on port ", s.Server.Addr)
	if err := s.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %v", err)
	}
	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) error {

	s.logger.Info("Shutting down http server on port ", s.Server.Addr)
	return s.Server.Shutdown(ctx)
}

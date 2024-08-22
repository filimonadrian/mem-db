package service

// import (
// 	"context"
// 	"fmt"
// 	config "mem-db/cmd/config"
// 	"net"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	// "google.golang.org/grpc/status"
// 	log "mem-db/cmd/logger"
// 	word "mem-db/pkg/proto/mem-db/pkg/proto/service"
// )

// type WSGrpcServer struct {
// 	word.UnimplementedWordServiceServer
// 	grpcServer *grpc.Server
// 	logger     log.Logger
// }

// func NewGrpcServer(ctx context.Context, options *config.ServiceOptions, svc *wordService) *WSGrpcServer {

// 	server := grpc.NewServer()
// 	word.RegisterWordServiceServer(server, svc)

// 	return &WSGrpcServer{
// 		grpcServer: server,
// 		logger:     ctx.Value(log.LoggerKey).(log.Logger),
// 	}
// }

// func (s *WSGrpcServer) Start() error {
// 	s.logger.Info("Starting server for WordService")

// 	lis, err := net.Listen("tcp", ":8080")
// 	if err != nil {
// 		return fmt.Errorf("Failed to listen on port 8080: %v", err)
// 	}

// 	return s.grpcServer.Serve(lis)
// }

// func (s *WSGrpcServer) Stop(ctx context.Context) error {
// 	s.grpcServer.GracefulStop()
// 	return nil
// }

// func (s *wordService) GetWordOccurences(ctx context.Context, req *word.GetWordOccurrencesRequest) (*word.GetWordOccurrencesResponse, error) {
// 	terms := req.GetTerms()
// 	if terms == "" {
// 		return &word.GetWordOccurrencesResponse{
// 			Status:     "Bad Request",
// 			StatusCode: int32(codes.InvalidArgument),
// 			Message:    "No words provided into request",
// 		}, nil
// 	}

// 	results := s.GetOccurences(terms)

// 	return &word.GetWordOccurrencesResponse{
// 		Status:     "Success",
// 		StatusCode: int32(codes.OK),
// 		Data:       results,
// 	}, nil
// }

// func (s *wordService) RegisterWordsG(ctx context.Context, req *word.RegisterWordsRequest) (*word.RegisterWordsResponse, error) {
// 	text := req.GetText()
// 	if text == "" {
// 		return &word.RegisterWordsResponse{
// 			Status:     "Bad Request",
// 			StatusCode: int32(codes.InvalidArgument),
// 			Message:    "Text field is empty",
// 		}, nil
// 	}

// 	s.RegisterWords(text)

// 	return &word.RegisterWordsResponse{
// 		Status:     "Success",
// 		StatusCode: int32(codes.OK),
// 		Message:    "Text processed successfully",
// 	}, nil
// }

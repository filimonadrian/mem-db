package service

import (
	"context"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	api "mem-db/pkg/api"
	repo "mem-db/pkg/repository"
	util "mem-db/pkg/util"
	"strings"
	"sync"
	"unicode"
)

type wordService struct {
	db           repo.DBService
	server       api.Server
	logger       log.Logger
	forwarding   bool
	forwardingCh chan []byte
}

type WordService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	SetForwarding()
	UnsetForwarding()
	SetForwardingCh(forwardingCh chan []byte)
}

type WordResponse struct {
	Word        string `json:"word"`
	Occurrences int    `json:"occurrences"`
}

func NewWordService(ctx context.Context, config *config.Config, db repo.DBService) WordService {
	ws := &wordService{
		db:     db,
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}

	ws.server = NewDBHttpServer(ctx, &config.ServiceOptions, ws)
	return ws
}

func (s *wordService) Start(ctx context.Context) error {
	return s.server.Start()
}

func (s *wordService) Stop(ctx context.Context) error {
	return s.server.Stop(ctx)
}

func (s *wordService) SetForwarding() {
	s.forwarding = true
}

func (s *wordService) UnsetForwarding() {
	s.forwarding = false
}

func (s *wordService) SetForwardingCh(forwardingCh chan []byte) {
	s.forwardingCh = forwardingCh
}

func (s *wordService) GetOccurences(terms string) []WordResponse {
	// terms = strings.ToLower(terms)
	words := strings.Split(terms, ",")

	var wg sync.WaitGroup
	wordOccChan := make(chan WordResponse, len(words))

	for _, word := range words {
		wg.Add(1)
		go func(word string) {
			defer wg.Done()

			word = strings.ToLower(word)
			occurrences := s.db.Get(word)

			// Send the result to the wordOccChan
			wordOccChan <- WordResponse{
				Word:        word,
				Occurrences: occurrences,
			}
		}(word)
	}

	go func() {
		wg.Wait()
		close(wordOccChan)
	}()

	var response []WordResponse

	for {
		select {
		case wordOcc, ok := <-wordOccChan:
			// channel is closed
			if !ok {
				return response
			}
			response = append(response, wordOcc)
		}
	}
	return response
}

func (s *wordService) RegisterWords(text string) {

	words := splitPhrase(text)

	wp := util.NewWorkerPool(15)
	wp.Start()

	// create a stream from words slice
	wordChannel := make(chan string)
	var wgStream sync.WaitGroup
	wgStream.Add(1)
	defer wgStream.Wait()

	go func() {
		wgStream.Done()
		for _, word := range words {
			wordChannel <- word
		}
		close(wordChannel)
	}()

	for word := range wordChannel {
		// avoid capturing loop variable
		word := strings.ToLower(word)
		wp.Submit(func() {
			s.db.Insert(word)
		})
	}

	wp.Stop()
}

func splitPhrase(text string) []string {
	splitter := func(c rune) bool {
		return unicode.IsSpace(c) || strings.ContainsRune(",.-_", c)
	}

	return strings.FieldsFunc(text, splitter)
}

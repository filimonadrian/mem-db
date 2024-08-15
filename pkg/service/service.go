package service

import (
	"context"
	"fmt"
	log "mem-db/cmd/logger"
	repo "mem-db/pkg/repository"
	"strings"
	"sync"
	"unicode"
)

type wordService struct {
	db     *repo.Database
	wal    *repo.WriteAheadLog
	logger log.Logger
}

type WordService interface {
	GetOccurences(terms string) []WordResponse
	RegisterWords(text string)
}

type WordResponse struct {
	Word        string `json:"word"`
	Occurrences int    `json:"occurrences"`
}

func NewService(ctx context.Context, db *repo.Database, wal *repo.WriteAheadLog) WordService {
	return &wordService{
		db:     db,
		wal:    wal,
		logger: ctx.Value(log.LoggerKey).(log.Logger),
	}
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

	wp := NewWorkerPool(5)
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
			err := s.wal.Write([]byte(word + "\n"))
			if err != nil {
				fmt.Println(err.Error())
			}
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

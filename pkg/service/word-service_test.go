package service

import (
	repo "mem-db/pkg/repository"
	"sync"
	"testing"
)

func Insert(dict *sync.Map, word string) {
	val, loaded := dict.LoadOrStore(word, 1)
	if loaded {
		dict.Store(word, val.(int)+1)
	}
}

func TestGetOccurences(t *testing.T) {

	db := &repo.Database{}
	mockDatastore := &sync.Map{}
	// Insert some words into the datastore
	Insert(mockDatastore, "apple")
	Insert(mockDatastore, "banana")
	Insert(mockDatastore, "banana")
	Insert(mockDatastore, "orange")
	Insert(mockDatastore, "orange")
	Insert(mockDatastore, "orange")

	db.SetDatastore(mockDatastore)

	ws := &wordService{
		db: db,
	}

	// Test GetOccurences function
	terms := "apple,banana,orange"
	result := ws.GetOccurences(terms)

	// Validate the results
	expected := map[string]int{
		"apple":  1,
		"banana": 2,
		"orange": 3,
	}

	resultMap := make(map[string]int)
	for _, res := range result {
		resultMap[res.Word] = res.Occurrences
	}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d results, but got %d", len(expected), len(result))
	}

	for word, occurrences := range expected {
		if resOccur, found := resultMap[word]; !found {
			t.Errorf("Expected word %v to be in the results", word)
		} else if resOccur != occurrences {
			t.Errorf("For word %v, expected %d occurrences but got %d", word, occurrences, resOccur)
		}
	}
}

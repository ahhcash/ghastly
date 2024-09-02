package inmemory

import (
	"errors"
	"fmt"
	"github.com/aakashshankar/vexdb/pkg/search"
	"sort"
	"sync"
)

type InMemoryStore struct {
	data map[string]VectorData
	mu   sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]VectorData),
	}
}

func (s *InMemoryStore) Get(key string) (*VectorData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if exists {
		return &value, nil
	}

	return nil, errors.New(fmt.Sprintf("key %s not found", key))
}

func (s *InMemoryStore) Put(key string, vector []float64, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = VectorData{
		vector: vector,
		meta:   metadata,
	}

	return nil
}

func (s *InMemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; exists {
		delete(s.data, key)
		return nil
	}

	return errors.New(fmt.Sprintf("key %s not found", key))
}

func (s *InMemoryStore) Search(queryVector []float64, topN int) ([]Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]Result, 0)

	for key, value := range s.data {
		score := search.Cosine(queryVector, value.vector)
		results = append(results, Result{
			Key:   key,
			Score: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results[:topN], nil
}

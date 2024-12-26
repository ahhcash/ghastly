package storage

import (
	"fmt"
	"github.com/ahhcash/ghastlydb/embed"
	"github.com/ahhcash/ghastlydb/search"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Result struct {
	Key   string
	Value string
	Score float64
}

type Store struct {
	memtable *Memtable
	sstables []*SSTable
	lock     sync.RWMutex
	destDir  string
	model    embed.Embedder
}

func NewStore(maxSize int, desDir string, model embed.Embedder) *Store {
	return &Store{
		memtable: NewMemtable(maxSize),
		sstables: []*SSTable{},
		destDir:  desDir,
		model:    model,
	}
}

func (s *Store) Put(key string, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	vector, err := s.model.Embed(value)
	if err != nil {
		return fmt.Errorf("could not embed Value %s: %v", value, err)
	}

	entry := Entry{
		Value:  value,
		Vector: vector,
	}
	err = s.memtable.Put(key, entry, s.destDir)
	if err != nil {
		return fmt.Errorf("could not Put data into memtable: %v", err)
	}

	// flushed to disk
	if s.memtable.Size() == 0 {
		err = s.loadNewSSTable()
		if err != nil {
			return fmt.Errorf("failed to load SSTable: %v", err)
		}
	}
	return nil
}

func (s *Store) Delete(key string) error {
	_, exists := s.Get(key)

	if !exists {
		return fmt.Errorf("key %s does not exist", key)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	tombstone := &Entry{
		Deleted:   true,
		Timestamp: time.Now().UnixMilli(),
	}

	err := s.memtable.Put(key, *tombstone, s.destDir)
	if err != nil {
		return fmt.Errorf("could not delete key %s: %v", key, err)
	}

	return nil
}

func (s *Store) Get(key string) (Entry, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	entry, exists := s.memtable.Get(key)
	if exists {
		if entry.Deleted {
			return Entry{}, false
		}
		return entry, true
	}

	for _, sstable := range s.sstables {
		entry, exists, _ = sstable.Get(key)

		if exists {
			return entry, true
		}

	}

	return Entry{}, false
}

func (s *Store) Search(query string, metric string) ([]Result, error) {
	queryVector, err := s.model.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("could not embed query vector: %v", err)
	}

	var scoreFn func([]float64, []float64) float64

	switch metric {
	case "dot":
		scoreFn = search.Dot
	case "l2":
		scoreFn = search.L2
	case "cosine":
		scoreFn = search.Cosine
	}

	results := make([]Result, 0)

	for _, sstable := range s.sstables {
		for _, key := range sstable.Index {
			entry, exists, err := sstable.Get(key)
			if err != nil {
				return nil, fmt.Errorf("could not fetch key %s from sstable: %v", key, err)
			}
			if exists {
				score := scoreFn(entry.Vector, queryVector)
				results = append(results, Result{
					Key:   key,
					Value: entry.Value,
					Score: score,
				})
			}
		}
	}

	// search memtable
	current := s.memtable.Data.head.next[0]
	for current != nil {
		entry, err := DeserializeEntry(current.value)
		if err != nil {
			fmt.Printf("could not deserialize entry: %v", err)
		}
		score := scoreFn(entry.Vector, queryVector)
		results = append(results, Result{
			Key:   current.key,
			Value: entry.Value,
			Score: score,
		})
		current = current.next[0]
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func (s *Store) Flush() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.memtable.Size() > 0 {
		err := s.memtable.flushToDisk(s.destDir)
		if err != nil {
			return fmt.Errorf("could not Flush memtable data to Disk: %v", err)
		}
		s.memtable.Clear()
		err = s.loadNewSSTable()
		if err != nil {
			return fmt.Errorf("could not load SSTable: %v", err)
		}
	}

	return nil
}

func (s *Store) loadNewSSTable() error {
	files, err := filepath.Glob(filepath.Join(s.destDir, "*.sst"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil // No SSTables to load
	}

	// Find the newest SSTable file (assuming filenames are UUIDs)
	newestFile := files[len(files)-1]

	// Open the SSTable
	sstable, err := OpenSSTable(newestFile)
	if err != nil {
		return err
	}

	// Prepend to the list of SSTables
	s.sstables = append([]*SSTable{sstable}, s.sstables...)

	return nil
}

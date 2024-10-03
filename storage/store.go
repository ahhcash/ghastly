package storage

import (
	"fmt"
	"path/filepath"
	"sync"
)

type Store struct {
	memtable *Memtable
	sstables []*SSTable
	lock     sync.RWMutex
	destDir  string
}

func NewStore(maxSize int, desDir string) *Store {
	return &Store{
		memtable: NewMemtable(maxSize),
		sstables: []*SSTable{},
		destDir:  desDir,
	}
}

func (s *Store) Put(key string, vector []float64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.memtable.Put(key, vector, s.destDir)
	if err != nil {
		return fmt.Errorf("could not Put data into memtable: %v", err)
	}

	if s.memtable.Size() == 0 {
		err = s.loadNewSSTable()
		if err != nil {
			return fmt.Errorf("failed to load SSTable: %v", err)
		}
	}
	return nil
}

func (s *Store) Get(key string) ([]float64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	vector, exists := s.memtable.Get(key)
	if exists {
		return vector, true
	}

	for _, sstable := range s.sstables {
		vector, exists, _ = sstable.Get(key)

		if exists {
			return vector, true
		}

	}

	return nil, false
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

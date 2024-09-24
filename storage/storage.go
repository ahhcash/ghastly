package storage

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Memtable struct {
	data    map[string][]byte
	lock    sync.RWMutex
	maxSize int
	size    int
}

func NewMemtable(maxSize int) *Memtable {
	return &Memtable{
		data:    make(map[string][]byte),
		maxSize: maxSize,
		size:    0,
	}
}

func (m *Memtable) Put(key string, value []byte, destPath string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, exists := m.data[key]; !exists {
		m.size++
	}
	m.data[key] = value

	if m.size >= m.maxSize {
		destPath = filepath.Join(destPath, uuid.New().String()+".pkl")
		err := m.FlushToDisk(destPath)
		if err != nil {
			panic(fmt.Errorf("could not flush memtable to disk: %v", err))
		}
		m.Clear()
	}
}

func (m *Memtable) Get(key string) ([]byte, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	value, exists := m.data[key]
	return value, exists
}

func (m *Memtable) Size() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.size
}

func (m *Memtable) FlushToDisk(destPath string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	err := os.MkdirAll(filepath.Dir(destPath), 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := m.data[key]
		_, err := file.WriteString(key + " : " + string(value) + "\n")
		if err != nil {
			return err
		}
	}

	m.data = make(map[string][]byte)
	m.size = 0

	return nil
}

func (m *Memtable) Clear() {
	m.data = make(map[string][]byte)
	m.size = 0
}

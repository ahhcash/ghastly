package storage

import (
	"fmt"
	"os"
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

func (m *Memtable) Put(key string, value []byte) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, exists := m.data[key]; !exists {
		m.size++
	}
	m.data[key] = value

	if m.size >= m.maxSize {
		m.flushToDisk()
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

func (m *Memtable) flushToDisk() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	err := os.MkdirAll("./data/", 0777)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("./data/sstable_%v", m.size)
	file, err := os.Create(fileName)
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

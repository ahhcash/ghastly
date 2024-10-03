package storage

import (
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"math"
	"os"
	"path/filepath"
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

func (m *Memtable) Put(key string, vector []float64, destPath string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var value []byte
	var err error

	if vector != nil {
		value, err = serializeVector(vector)
		if err != nil {
			return fmt.Errorf("error when serializing data: %v", err)
		}
	}

	_, exists := m.data[key]
	if !exists {
		m.size++
	}

	if err != nil {
		return fmt.Errorf("could not write data into memtable: %v", err)
	}

	m.data[key] = value

	if m.size >= m.maxSize {
		err := m.flushToDisk(destPath)
		if err != nil {
			return fmt.Errorf("could not flush memtable to disk: %v", err)
		}
		m.clear()
	}

	return nil
}

func serializeVector(vector []float64) ([]byte, error) {
	buf := make([]byte, 8*len(vector))
	for i, v := range vector {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return buf, nil
}

func (m *Memtable) Get(key string) ([]float64, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	value, exists := m.data[key]
	if !exists {
		return nil, false
	}

	vector, err := deserializeVector(value)
	if err != nil {
		panic(fmt.Errorf("could not deserialize vector: %v", err))
	}

	return vector, exists
}

func deserializeVector(data []byte) ([]float64, error) {
	vector := make([]float64, len(data)/8)
	for i := range vector {
		bits := binary.LittleEndian.Uint64(data[i*8:])
		vector[i] = math.Float64frombits(bits)
	}

	return vector, nil
}

func (m *Memtable) Size() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.size
}

func (m *Memtable) flushToDisk(destPath string) error {
	err := os.MkdirAll(filepath.Dir(destPath), 0777)
	if err != nil {
		return fmt.Errorf("could not make path for %s: %v", destPath, err)
	}
	filename := filepath.Join(destPath, uuid.New().String()+".sst")
	tempFilename := filename + ".tmp"

	file, err := os.Create(tempFilename)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", tempFilename, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	for key, val := range m.data {
		err := writeRecord(file, key, val)
		if err != nil {
			return fmt.Errorf("could not write record: %v", err)
		}
	}

	err = os.Rename(tempFilename, filename)
	if err != nil {
		return fmt.Errorf("could not rename temp file to sst: %v", err)
	}

	return nil
}

func writeRecord(file *os.File, key string, value []byte) error {
	err := binary.Write(file, binary.LittleEndian, int32(len(key)))
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, int32(len(value)))
	if err != nil {
		return err
	}

	_, err = file.Write(value)
	if err != nil {
		return err
	}

	return nil
}

func (m *Memtable) clear() {
	m.data = make(map[string][]byte)
	m.size = 0
}

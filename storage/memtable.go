package storage

import (
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"math"
	"os"
	"path/filepath"
)

type Entry struct {
	Value     string
	Vector    []float64
	Deleted   bool
	Timestamp int64
}

type Memtable struct {
	Data    *SkipList
	maxSize int
	size    int
}

func NewMemtable(maxSize int) *Memtable {
	return &Memtable{
		Data:    NewSkipList(),
		maxSize: maxSize,
		size:    0,
	}
}

func (m *Memtable) Put(key string, entry Entry, destPath string) error {
	var value []byte
	var err error

	value, err = SerializeEntry(entry)
	if err != nil {
		return fmt.Errorf("error when serializing data: %v", err)
	}

	_, exists := m.Data.Search(key)
	if !exists {
		m.size++
	}

	m.Data.Insert(key, value)

	if m.size >= m.maxSize {
		err := m.flushToDisk(destPath)
		if err != nil {
			return fmt.Errorf("could not flush memtable to disk: %v", err)
		}
		m.Clear()
	}

	return nil
}

func SerializeEntry(entry Entry) ([]byte, error) {
	valueLen := int32(len(entry.Value))
	vectorLen := int32(len(entry.Vector))
	totalBufSize := 1 + 8 + 4 + valueLen + 4 + 8*vectorLen
	buf := make([]byte, totalBufSize)
	offset := 0

	// write deleted fisrt
	if entry.Deleted {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
	offset += 1

	binary.LittleEndian.PutUint64(buf[offset:], uint64(entry.Timestamp))
	offset += 8

	binary.LittleEndian.PutUint32(buf[offset:], uint32(valueLen))
	offset += 4

	copy(buf[offset:], entry.Value)
	offset += int(valueLen)

	binary.LittleEndian.PutUint32(buf[offset:], uint32(vectorLen))
	offset += 4

	for i, v := range entry.Vector {
		binary.LittleEndian.PutUint64(buf[offset+i*8:], math.Float64bits(v))
	}

	return buf, nil
}

func (m *Memtable) Get(key string) (Entry, bool) {
	value, exists := m.Data.Search(key)
	if !exists {
		return Entry{}, false
	}

	entry, err := DeserializeEntry(value)
	if err != nil {
		panic(fmt.Errorf("could not deserialize vector: %v", err))
	}

	return entry, exists
}

func DeserializeEntry(data []byte) (Entry, error) {
	// 1 + 8 + 4 + 4
	if len(data) < 17 {
		return Entry{}, fmt.Errorf("data insufficent to deserialize, got %d bytes", len(data))
	}

	var offset = 0
	deleted := data[offset] == 1
	offset += 1

	timestamp := int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8

	valueLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	if offset+int(valueLen) > len(data) {
		return Entry{}, fmt.Errorf("invalid value length, reading past end of data")
	}

	value := data[offset : offset+int(valueLen)]
	offset += int(valueLen)

	if offset+4 > len(data) {
		return Entry{}, fmt.Errorf("invalid vector length: no space for vector")
	}

	vectorLen := binary.LittleEndian.Uint32(data[offset:])
	offset += int(vectorLen)

	requiredBytes := offset + int(vectorLen)*8
	if requiredBytes > len(data) {
		return Entry{}, fmt.Errorf("invalid vector length: reading past end of data")
	}
	vector := make([]float64, vectorLen)
	for i := range vector {
		bits := binary.LittleEndian.Uint64(data[offset:])
		vector[i] = math.Float64frombits(bits)
		offset += 8
	}

	return Entry{
		Value:     string(value),
		Vector:    vector,
		Deleted:   deleted,
		Timestamp: timestamp,
	}, nil
}

func (m *Memtable) Size() int {
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
			return
		}
	}(file)
	current := m.Data.head.next[0]
	for current != nil {
		err := writeRecord(file, current.key, current.value)
		if err != nil {
			return fmt.Errorf("could not write record: %v", err)
		}
		current = current.next[0]
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

func (m *Memtable) Clear() {
	m.Data = NewSkipList()
	m.size = 0
}

package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
)

type SSTable struct {
	file      *os.File
	index     []string
	positions []int64
}

func OpenSSTable(filename string) (*SSTable, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open sstable %s: %v", filename, err)
	}

	sst := &SSTable{
		file:      file,
		index:     []string{},
		positions: []int64{},
	}

	err = sst.buildIndex()
	if err != nil {
		return nil, fmt.Errorf("could not build sst index: %v", err)
	}
	return sst, nil
}

func (sst *SSTable) buildIndex() error {
	var offset int64 = 0

	for {
		var keyLen int32
		err := binary.Read(sst.file, binary.LittleEndian, &keyLen)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		keyBytes := make([]byte, keyLen)
		_, err = io.ReadFull(sst.file, keyBytes)
		if err != nil {
			return err
		}
		key := string(keyBytes)

		var valueLen int32
		err = binary.Read(sst.file, binary.LittleEndian, &valueLen)
		if err != nil {
			return err
		}

		_, err = sst.file.Seek(int64(valueLen), io.SeekCurrent)
		if err != nil {
			return err
		}

		sst.index = append(sst.index, key)
		sst.positions = append(sst.positions, offset)

		offset, err = sst.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sst *SSTable) Close() error {
	return sst.file.Close()
}

func (sst *SSTable) Get(key string) ([]float64, bool, error) {
	i := sort.SearchStrings(sst.index, key)
	if i >= len(sst.index) || sst.index[i] != key {
		return nil, false, nil
	}

	_, err := sst.file.Seek(sst.positions[i], io.SeekStart)
	if err != nil {
		return nil, false, fmt.Errorf("could not seek within the file: %v", err)
	}

	var keyLen int32
	err = binary.Read(sst.file, binary.LittleEndian, &keyLen)
	if err != nil {
		return nil, false, fmt.Errorf("could not read key length: %v", err)
	}

	keyBytes := make([]byte, keyLen)
	_, err = io.ReadFull(sst.file, keyBytes)
	if err != nil {
		return nil, false, fmt.Errorf("could not read key bytes: %v", err)
	}

	var valueLen int32
	err = binary.Read(sst.file, binary.LittleEndian, &valueLen)
	if err != nil {
		return nil, false, fmt.Errorf("could not read value length: %v", err)
	}

	valueBytes := make([]byte, valueLen)
	_, err = io.ReadFull(sst.file, valueBytes)
	if err != nil {
		return nil, false, fmt.Errorf("could not read value bytes: %v", err)
	}

	vector, err := deserializeVector(valueBytes)
	if err != nil {
		return nil, false, fmt.Errorf("could not deserialize value bytes: %v", err)
	}

	return vector, true, nil
}

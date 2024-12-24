package tests

import (
	"github.com/ahhcash/ghastlydb/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

type MemtableTestSuite struct {
	suite.Suite
	memtable *storage.Memtable
	testPath string
}

func (s *MemtableTestSuite) SetupTest() {
	s.memtable = storage.NewMemtable(1024)
	s.testPath = "./test_data"
	_ = os.MkdirAll(s.testPath, 0755)
}

func (s *MemtableTestSuite) TearDownTest() {
	_ = os.RemoveAll(s.testPath)
}

func (s *MemtableTestSuite) TestNewMemtable() {
	assert.NotNil(s.T(), s.memtable)
	assert.Equal(s.T(), 0, s.memtable.Size())
}

func (s *MemtableTestSuite) TestPutAndGet() {
	entry := storage.Entry{
		Value:  "test value",
		Vector: []float64{1.0, 2.0, 3.0},
	}

	// Test Put
	err := s.memtable.Put("test_key", entry, s.testPath)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, s.memtable.Size())

	// Test Get
	retrieved, exists := s.memtable.Get("test_key")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), entry.Value, retrieved.Value)
	assert.Equal(s.T(), entry.Vector, retrieved.Vector)

	// Test non-existent key
	_, exists = s.memtable.Get("nonexistent")
	assert.False(s.T(), exists)
}

func (s *MemtableTestSuite) TestUpdateExisting() {
	entry1 := storage.Entry{
		Value:  "initial value",
		Vector: []float64{1.0, 2.0},
	}
	entry2 := storage.Entry{
		Value:  "updated value",
		Vector: []float64{3.0, 4.0},
	}

	// Insert initial entry
	err := s.memtable.Put("key", entry1, s.testPath)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, s.memtable.Size())

	// Update with new entry
	err = s.memtable.Put("key", entry2, s.testPath)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, s.memtable.Size()) // Size shouldn't change

	// Verify update
	retrieved, exists := s.memtable.Get("key")
	assert.True(s.T(), exists)
	assert.Equal(s.T(), entry2.Value, retrieved.Value)
	assert.Equal(s.T(), entry2.Vector, retrieved.Vector)
}

func (s *MemtableTestSuite) TestFlushToDisk() {
	smallMemtable := storage.NewMemtable(32) // Small size to force flush

	// Add entries until flush
	for i := 0; i < 100; i++ {
		entry := storage.Entry{
			Value:  "test value",
			Vector: []float64{1.0, 2.0},
		}
		err := smallMemtable.Put("key"+string(rune(i)), entry, s.testPath)
		assert.NoError(s.T(), err)
	}

	// Verify SST file was created
	files, err := os.ReadDir(s.testPath)
	assert.NoError(s.T(), err)
	assert.True(s.T(), len(files) > 0)

	for _, file := range files {
		assert.Equal(s.T(), ".sst", filepath.Ext(file.Name()))
	}
}

func (s *MemtableTestSuite) TestClear() {
	entry := storage.Entry{
		Value:  "test value",
		Vector: []float64{1.0, 2.0},
	}

	// Add some data
	err := s.memtable.Put("key", entry, s.testPath)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, s.memtable.Size())

	// Clear memtable
	s.memtable.Clear()
	assert.Equal(s.T(), 0, s.memtable.Size())

	// Verify data is gone
	_, exists := s.memtable.Get("key")
	assert.False(s.T(), exists)
}

func (s *MemtableTestSuite) TestSerializeDeserialize() {
	original := storage.Entry{
		Value:  "test value",
		Vector: []float64{1.0, 2.0, 3.0},
	}

	// Serialize
	serialized, err := storage.SerializeEntry(original)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), serialized)

	// Deserialize
	deserialized, err := storage.DeserializeEntry(serialized)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), original.Value, deserialized.Value)
	assert.Equal(s.T(), original.Vector, deserialized.Vector)
}

func TestMemtableSuite(t *testing.T) {
	suite.Run(t, new(MemtableTestSuite))
}

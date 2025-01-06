package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SkipListTestSuite struct {
	suite.Suite
	list *SkipList
}

func (s *SkipListTestSuite) SetupTest() {
	s.list = NewSkipList()
}

// Test initialization
func (s *SkipListTestSuite) TestNewSkipList() {
	// Test initialization
	assert.NotNil(s.T(), s.list, "Skiplist should not be nil")
	assert.Equal(s.T(), 0, s.list.level, "Initial level should be 0")
	assert.Equal(s.T(), 0, s.list.length, "Initial length should be 0")
	assert.NotNil(s.T(), s.list.head, "head node should not be nil")
	assert.Equal(s.T(), maxLevel, len(s.list.head.next),
		"head node should have maxLevel next pointers")
}

// Test single insert and search
func (s *SkipListTestSuite) TestBasicInsertAndSearch() {
	// Test insert
	testKey := "test_key"
	testValue := []byte("test_value")
	s.list.Insert(testKey, testValue)

	// Verify length increased
	assert.Equal(s.T(), 1, s.list.length)

	// Test successful search
	value, exists := s.list.Search(testKey)
	assert.True(s.T(), exists)
	assert.Equal(s.T(), testValue, value)

	// Test search for non-existent key
	value, exists = s.list.Search("nonexistent")
	assert.False(s.T(), exists)
	assert.Nil(s.T(), value)
}

// Test updating existing key
func (s *SkipListTestSuite) TestUpdate() {
	testKey := "update_key"
	originalValue := []byte("original")
	updatedValue := []byte("updated")

	// Insert original value
	s.list.Insert(testKey, originalValue)
	assert.Equal(s.T(), 1, s.list.length)

	// Update with new value
	s.list.Insert(testKey, updatedValue)
	assert.Equal(s.T(), 1, s.list.length) // length should not change

	// Verify update
	value, exists := s.list.Search(testKey)
	assert.True(s.T(), exists)
	assert.Equal(s.T(), updatedValue, value)
}

// Test empty list operations
func (s *SkipListTestSuite) TestEmptyList() {
	value, exists := s.list.Search("any_key")
	assert.False(s.T(), exists)
	assert.Nil(s.T(), value)
	assert.Equal(s.T(), 0, s.list.length)
}

// Test iterator on empty list
func (s *SkipListTestSuite) TestEmptyIterator() {
	iter := s.list.Iterator()
	_, _, hasNext := iter()
	assert.False(s.T(), hasNext)
}

func (s *SkipListTestSuite) TestBasicIterator() {
	testData := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	for k, v := range testData {
		s.list.Insert(k, v)
	}

	iter := s.list.Iterator()
	count := 0
	for key, value, hasNext := iter(); hasNext; key, value, hasNext = iter() {
		count++
		expectedValue, exists := testData[key]
		assert.True(s.T(), exists)
		assert.Equal(s.T(), expectedValue, value)
	}

	assert.Equal(s.T(), len(testData), count)
}

func TestSkipList(t *testing.T) {
	suite.Run(t, new(SkipListTestSuite))
}

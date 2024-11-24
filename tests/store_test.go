package tests

import (
	"fmt"
	"github.com/aakashshankar/vexdb/storage"
	"github.com/aakashshankar/vexdb/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"os"
	"sync"
	"testing"
)

type StoreTestSuite struct {
	suite.Suite
	testDestDir string
	store       *storage.Store
	emb         *mocks.MockEmbedder
}

func (s *StoreTestSuite) SetupSuite() {
	_ = os.MkdirAll(s.testDestDir, 0777)
}

func (s *StoreTestSuite) SetupTest() {
	s.emb = new(mocks.MockEmbedder)
	s.store = storage.NewStore(64, s.testDestDir, s.emb)
	s.emb.On("Embed", mock.AnythingOfType("string")).Return(
		[]float64{0.1, 0.2, 0.3213},
		nil,
	)
}

func (s *StoreTestSuite) TearDownSuite() {
	_ = os.RemoveAll(s.testDestDir)
}

func (s *StoreTestSuite) TestPut() {
	err := s.store.Put("test_key", "test_value")

	assert.NoError(s.T(), err)
	s.emb.AssertCalled(s.T(), "Embed", "test_value")
	s.emb.AssertNumberOfCalls(s.T(), "Embed", 1)
}

func (s *StoreTestSuite) TestMultiThreadedPut() {
	numRoutines := 10
	var wg sync.WaitGroup
	wg.Add(numRoutines)

	errs := make(chan error, numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			val := fmt.Sprintf("val-%d", i)
			if err := s.store.Put(key, val); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		assert.NoError(s.T(), err)
	}

	s.emb.AssertNumberOfCalls(s.T(), "Embed", numRoutines)
}

func (s *StoreTestSuite) TestGet() {
	_ = s.store.Put("test-key", "test-value")
	entry, exists := s.store.Get("test-key")

	assert.Equal(s.T(), entry.Value, "test-value")
	assert.True(s.T(), exists)

	entry, exists = s.store.Get("blah")
	assert.False(s.T(), exists)
}

func (s *StoreTestSuite) TestMultiThreadedGet() {
	numRoutines := 10
	var wg sync.WaitGroup
	wg.Add(numRoutines)

	_ = s.store.Put("test-key", "test-val")

	errs := make(chan bool, numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			if _, exists := s.store.Get("test-key"); !exists {
				errs <- exists
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	assert.Empty(s.T(), errs)
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

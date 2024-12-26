package tests

import (
	"github.com/ahhcash/ghastlydb/db"
	"github.com/ahhcash/ghastlydb/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type DBTestSuite struct {
	suite.Suite
	testPath string
}

func (s *DBTestSuite) SetupSuite() {
	_ = os.Setenv("OPENAI_API_KEY", "test-key")
	_ = os.Setenv("NV_API_KEY", "test-key")
}

func (s *DBTestSuite) TearDownSuite() {
	_ = os.Unsetenv("OPENAI_API_KEY")
	_ = os.Unsetenv("NV_API_KEY")
}

func (s *DBTestSuite) SetupTest() {
	s.testPath = "./test_db"
}

func (s *DBTestSuite) TearDownTest() {
	_ = os.RemoveAll(s.testPath)
}

func (s *DBTestSuite) TestDefaultConfig() {
	cfg := db.DefaultConfig()

	assert.Equal(s.T(), "./ghastlydb_data", cfg.Path)
	assert.Equal(s.T(), "colbert", cfg.EmbeddingModel)
	assert.Equal(s.T(), 64*1024*1024, cfg.MemtableSize)
	assert.Equal(s.T(), "cosine", cfg.Metric)
}

func (s *DBTestSuite) TestOpenDB() {
	cfg := db.Config{
		Metric:         "dot",
		EmbeddingModel: "nvidia",
		MemtableSize:   1024,
		Path:           s.testPath,
	}

	database, err := db.OpenDB(cfg)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), database)
}

func (s *DBTestSuite) TestOpenDBWithEmbedder() {
	mockEmbedder := mocks.MockEmbedder{}
	mockEmbedder.On("Embed", mock.AnythingOfType("string")).Return(
		[]float64{0.1, 0.2, 0.3213},
		nil,
	)

	cfg := db.Config{
		Metric:         "dot",
		EmbeddingModel: "nvidia",
		MemtableSize:   1024,
		Path:           s.testPath,
	}

	db2, err := db.OpenDBWithEmbedder(cfg, &mockEmbedder)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), db2)
}

func (s *DBTestSuite) TestPutAndGet() {
	mockEmbedder := mocks.MockEmbedder{}
	mockEmbedder.On("Embed", mock.AnythingOfType("string")).Return(
		[]float64{0.342323, 0.556455, 0.43244},
		nil,
	)

	cfg := db.Config{
		Path:           s.testPath,
		MemtableSize:   1024,
		Metric:         "cosine",
		EmbeddingModel: "openai",
	}

	database, err := db.OpenDBWithEmbedder(cfg, &mockEmbedder)
	require.NoError(s.T(), err)

	err = database.Put("test_key", "test_value")
	assert.NoError(s.T(), err)

	value, err := database.Get("test_key")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test_value", value)

	_, err = database.Get("non_existent_key")
	assert.Error(s.T(), err)
}

func (s *DBTestSuite) TestDelete() {
	mockEmbedder := mocks.MockEmbedder{}
	mockEmbedder.On("Embed", mock.AnythingOfType("string")).Return(
		[]float64{0.342323, 0.556455, 0.43244},
		nil,
	)

	cfg := db.Config{
		Path:           s.testPath,
		MemtableSize:   1024,
		Metric:         "cosine",
		EmbeddingModel: "openai",
	}

	database, err := db.OpenDBWithEmbedder(cfg, &mockEmbedder)
	require.NoError(s.T(), err)

	err = database.Put("test_key", "test_value")
	assert.NoError(s.T(), err)

	err = database.Delete("test_key")
	assert.NoError(s.T(), err)
	//
	//_, err = database.Get("test_key")
	//assert.Error(s.T(), err)
}

func (s *DBTestSuite) TestSearch() {
	mockEmbedder := mocks.MockEmbedder{}
	mockEmbedder.On("Embed", mock.AnythingOfType("string")).Return(
		[]float64{0.43324, 0.4324532, 0.432424},
		nil,
	)

	cfg := db.Config{
		Path:           s.testPath,
		MemtableSize:   1024,
		Metric:         "cosine",
		EmbeddingModel: "openai",
	}

	database, err := db.OpenDBWithEmbedder(cfg, &mockEmbedder)
	require.NoError(s.T(), err)

	testData := map[string]string{
		"key1": "This is a test document",
		"key2": "Another test document",
	}

	for k, v := range testData {
		err := database.Put(k, v)
		require.NoError(s.T(), err)
	}

	results, err := database.Search("test document")
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), results)
}

func TestDBSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

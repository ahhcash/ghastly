package db

import (
	"fmt"
	"github.com/ahhcash/ghastlydb/embed"
	"github.com/ahhcash/ghastlydb/embed/local/colbert"
	"github.com/ahhcash/ghastlydb/embed/nvidia"
	"github.com/ahhcash/ghastlydb/embed/openai"
	"github.com/ahhcash/ghastlydb/storage"
	"os"
)

type Config struct {
	Path           string
	MemtableSize   int
	Metric         string
	EmbeddingModel string
}

type DB struct {
	store  *storage.Store
	config Config
}

func initializeEmbeddingModel(model string) (embed.Embedder, error) {
	switch model {
	case "openai":
		return openai.NewOpenAIEmbedder()
	case "nvidia":
		return nvidia.LoadNvidiaEmbedder()
	case "colbert":
		return colbert.NewColBERTEmbedder()
	default:
		return nil, fmt.Errorf("embedding model %s not supported", model)
	}
}

func DefaultConfig() Config {
	return Config{
		Path:           "./ghastlydb_data",
		MemtableSize:   64 * 1024 * 1024,
		Metric:         "cosine",
		EmbeddingModel: "colbert",
	}
}

func OpenDB(cfg Config) (*DB, error) {
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		return nil, fmt.Errorf("could not create db directory at %s: %v", cfg.Path, err)
	}
	model, err := initializeEmbeddingModel(cfg.EmbeddingModel)
	if err != nil {
		fmt.Printf("could not initialize embedding model: %v", err)
		os.Exit(1)
	}

	store := storage.NewStore(cfg.MemtableSize, cfg.Path, model)

	return &DB{
		store:  store,
		config: cfg,
	}, nil
}

func OpenDBWithEmbedder(cfg Config, embedder embed.Embedder) (*DB, error) {
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		return nil, fmt.Errorf("could not create db directory at %s: %v", cfg.Path, err)
	}
	store := storage.NewStore(cfg.MemtableSize, cfg.Path, embedder)

	return &DB{
		store:  store,
		config: cfg,
	}, nil
}

func (db *DB) Put(key string, value string) error {
	return db.store.Put(key, value)
}

func (db *DB) Get(key string) (string, error) {
	entry, exists := db.store.Get(key)
	if !exists {
		return "", fmt.Errorf("key %s does not exist\n", key)
	}

	return entry.Value, nil
}

func (db *DB) Search(query string) ([]storage.Result, error) {
	return db.store.Search(query, db.config.Metric)
}

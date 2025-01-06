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

type DBConfig struct {
	Path           string
	MemtableSize   int
	Metric         string
	EmbeddingModel string
}

type DB struct {
	store    *storage.Store
	DBConfig DBConfig
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

func DefaultConfig() DBConfig {
	return DBConfig{
		Path:           "./ghastlydb_data",
		MemtableSize:   64 * 1024 * 1024,
		Metric:         "cosine",
		EmbeddingModel: "openai",
	}
}

func OpenDB(cfg DBConfig) (*DB, error) {
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
		store:    store,
		DBConfig: cfg,
	}, nil
}

func OpenDBWithEmbedder(cfg DBConfig, embedder embed.Embedder) (*DB, error) {
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		return nil, fmt.Errorf("could not create db directory at %s: %v", cfg.Path, err)
	}
	store := storage.NewStore(cfg.MemtableSize, cfg.Path, embedder)

	return &DB{
		store:    store,
		DBConfig: cfg,
	}, nil
}

func (db *DB) Put(key string, value string) error {
	return db.store.Put(key, value)
}

func (db *DB) Delete(key string) error {
	return db.store.Delete(key)
}

func (db *DB) Get(key string) (string, error) {
	entry, exists := db.store.Get(key)
	if !exists {
		return "", fmt.Errorf("key %s does not exist\n", key)
	}

	return entry.Value, nil
}

func (db *DB) Exists(key string) bool {
	_, exists := db.store.Get(key)
	return exists
}

func (db *DB) Search(query string) ([]storage.Result, error) {
	return db.store.Search(query, db.DBConfig.Metric)
}

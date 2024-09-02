package inmemory

type Store interface {
	Put(key string, vector []float64, metadata map[string]string) error

	Get(key string) (*VectorData, error)

	Delete(key string) error
}

type VectorData struct {
	vector []float64
	meta   map[string]string
}

type Result struct {
	Key   string
	Score float64
}

package embed

type Embedder interface {
	Embed(text string) ([]float64, error)
}

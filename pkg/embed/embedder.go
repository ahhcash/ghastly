package embed

type embedder interface {
	Embed(text string) ([]float64, error)
}

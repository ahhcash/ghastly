package colbert

import (
	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/options"
	"github.com/knights-analytics/hugot/pipelines"
	"os"
	"path/filepath"
)

type ColBERTEmbedder struct {
	pipeline *pipelines.FeatureExtractionPipeline
}

func NewColBERTEmbedder() (*ColBERTEmbedder, error) {
	osConfig := getConfig()
	onnxPath := osConfig.OnnxPath()
	session, err := hugot.NewORTSession(
		options.WithOnnxLibraryPath(onnxPath))
	if err != nil {
		return nil, err
	}

	pwd, _ := os.Getwd()
	dest := filepath.Join(pwd, "models")

	modelPath, err := hugot.DownloadModel(modelHfPath, dest, hugot.NewDownloadOptions())
	if err != nil {
		return nil, err
	}
	pipelineConfig := hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
		Name:      "EmbeddingPipeline",
	}

	embeddingPipeline, err := hugot.NewPipeline(session, pipelineConfig)
	if err != nil {
		return nil, err
	}

	return &ColBERTEmbedder{
		pipeline: embeddingPipeline,
	}, nil
}

func (c *ColBERTEmbedder) Embed(text string) ([]float64, error) {
	batch := []string{text}
	result, err := c.pipeline.RunPipeline(batch)
	if err != nil {
		return nil, err
	}
	embeddings := make([]float64, len(result.Embeddings[0]))
	for i, v := range result.Embeddings[0] {
		embeddings[i] = float64(v)
	}
	return embeddings, nil
}

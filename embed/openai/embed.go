package openai

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"os"
)

const (
	model      = "text-embedding-3-small"
	apiBaseUrl = "https://api.openai.com/v1/embeddings"
)

type OpenAIEmbedder struct {
	apiKey     string
	apiBaseUrl string
}

func NewOpenAIEmbedder() (*OpenAIEmbedder, error) {
	apiKey, exists := os.LookupEnv("OPENAI_API_KEY")
	if !exists {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &OpenAIEmbedder{
		apiKey:     apiKey,
		apiBaseUrl: apiBaseUrl,
	}, nil
}

func (e *OpenAIEmbedder) GetEmbeddings(input string) (*OpenAIEmbeddingResponse, error) {
	jsonBody, err := marshalRequest(input)
	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(e.apiBaseUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.SetBody(jsonBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode(), resp.Body())
	}

	var embedResponse OpenAIEmbeddingResponse
	err = json.Unmarshal(resp.Body(), &embedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &embedResponse, nil
}

func marshalRequest(input string) ([]byte, error) {
	requestBody := OpenAIEmbeddingRequest{
		Input: []string{input},
		Model: model,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return jsonBody, nil
}

func (e *OpenAIEmbedder) Embed(text string) ([]float64, error) {
	response, err := e.GetEmbeddings(text)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// Convert []float32 to []float64
	embedding := make([]float64, len(response.Data[0].Embedding))
	for i, v := range response.Data[0].Embedding {
		embedding[i] = float64(v)
	}

	return embedding, nil
}

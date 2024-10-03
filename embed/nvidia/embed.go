package nvidia

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"os"
)

const (
	model      = "nvidia/nv-embedqa-mistral-7b-v2"
	apiBaseUrl = "https://integrate.api.nvidia.com"
)

type NvidiaEmbedder struct {
	apiBaseUrl string
	apiKey     string
}

func LoadNvidiaEmbedder() (*NvidiaEmbedder, error) {
	apiKey, exists := os.LookupEnv("NV_API_KEY")
	if !exists {
		return nil, fmt.Errorf("NV_API_KEY not set")
	}

	return &NvidiaEmbedder{
		apiBaseUrl: apiBaseUrl,
		apiKey:     apiKey,
	}, nil
}

func (nv *NvidiaEmbedder) GetEmbeddings(input string) (*NVEmbeddingResponse, error) {
	jsonBody, err2 := marshalRequest(input)
	if err2 != nil {
		return nil, err2
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(nv.apiBaseUrl + "v1/embeddings")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", "Bearer "+nv.apiKey)
	req.SetBody(jsonBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("failed to get embeddings: %d, %s", resp.StatusCode(), resp.Body())
	}

	var nvResp NVEmbeddingResponse
	err := json.Unmarshal(resp.Body(), &nvResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response %w", err)
	}

	return &nvResp, nil
}

func marshalRequest(input string) ([]byte, error) {
	reqBody := NVEmbeddingRequest{
		Input:          []string{input},
		Model:          model,
		InputType:      "query",
		EncodingFormat: "float",
		Truncate:       "NONE",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func (nv *NvidiaEmbedder) Embed(text string) ([]float64, error) {
	nvResp, err := nv.GetEmbeddings(text)
	if err != nil {
		return nil, err
	}

	return nvResp.Data[0].Embedding, nil
}

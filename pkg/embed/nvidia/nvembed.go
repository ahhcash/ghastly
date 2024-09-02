package nvidia

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
)

const (
	model = "nvidia/nv-embedqa-mistral-7b-v2"
)

var nvApiConfig *Config

func GetNVEmbeddings(input []string) (*NVEmbeddingResponse, error) {
	jsonBody, err2 := marshalRequest(input)
	if err2 != nil {
		return nil, err2
	}

	nvApiConfig, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(nvApiConfig.apiBaseUrl + "v1/embeddings")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", "Bearer "+nvApiConfig.apiKey)
	req.SetBody(jsonBody)

	if err := fasthttp.Do(req, resp); err != nil {
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("failed to get embeddings: %d, %s", resp.StatusCode(), resp.Body())
	}

	var nvResp NVEmbeddingResponse
	err = json.Unmarshal(resp.Body(), &nvResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding response %w", err)
	}

	return &nvResp, nil
}

func marshalRequest(input []string) ([]byte, error) {
	reqBody := NVEmbeddingRequest{
		Input:          input,
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

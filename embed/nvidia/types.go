package nvidia

type NVEmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	InputType      string   `json:"input_type"`
	EncodingFormat string   `json:"encoding_format"`
	Truncate       string   `json:"truncate"`
}

type NVEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

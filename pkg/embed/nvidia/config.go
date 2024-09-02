package nvidia

import (
	"fmt"
	"os"
)

type Config struct {
	apiBaseUrl string
	apiKey     string
}

func LoadConfig() (*Config, error) {
	apiKey, exists := os.LookupEnv("NV_API_KEY")
	if !exists {
		return nil, fmt.Errorf("NV_API_KEY not set")
	}

	return &Config{
		apiBaseUrl: "https://integrate.api.nvidia.com/",
		apiKey:     apiKey,
	}, nil
}

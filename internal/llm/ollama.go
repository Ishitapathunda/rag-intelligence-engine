package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaLLM struct {
	model string
	url   string
}

func NewOllama(baseURL, model string) *OllamaLLM {
	return &OllamaLLM{
		model: model,
		url:   baseURL + "/api/generate",
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
}

func (o *OllamaLLM) Generate(prompt string) (string, error) {
	body, err := json.Marshal(generateRequest{Model: o.model, Prompt: prompt, Stream: false})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(o.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama generate failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned %d", resp.StatusCode)
	}

	var parsed generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	return parsed.Response, nil
}

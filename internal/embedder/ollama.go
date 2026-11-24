package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaEmbedder struct {
	model string
	url   string
}

func NewOllama(model string) *OllamaEmbedder {
	return &OllamaEmbedder{
		model: model,
		url:   "http://localhost:11434/api/embed",
	}
}

type ollamaRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (o *OllamaEmbedder) Embed(texts []string) ([][]float32, error) {
	body, err := json.Marshal(ollamaRequest{Model: o.model, Input: texts})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(o.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama request failed (is `ollama serve` running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned %d", resp.StatusCode)
	}

	var parsed ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return parsed.Embeddings, nil
}

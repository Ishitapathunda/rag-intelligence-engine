package config

import (
	"os"
	"strconv"
)

type Config struct {
	OllamaURL      string
	EmbeddingModel string
	LLMModel       string
	QdrantHost     string
	QdrantPort     int
	CollectionName string
	TopK           uint64
	ServerAddress  string

	RawDir       string
	ProcessedDir string
	ChunkSize    int
	ChunkOverlap int
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func getEnvUint(key string, fallback uint64) uint64 {
	value, err := strconv.ParseUint(os.Getenv(key), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func Load() Config {
	return Config{
		OllamaURL:      getEnv("OLLAMA_URL", "http://localhost:11434"),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", "nomic-embed-text"),
		LLMModel:       getEnv("LLM_MODEL", "llama3.1"),
		QdrantHost:     getEnv("QDRANT_HOST", "localhost"),
		QdrantPort:     getEnvInt("QDRANT_PORT", 6334),
		CollectionName: getEnv("COLLECTION_NAME", "distsys-corpus"),
		TopK:           getEnvUint("TOP_K", 5),
		ServerAddress:  getEnv("SERVER_ADDRESS", ":8080"),

		RawDir:       getEnv("RAW_DIR", "data/raw"),
		ProcessedDir: getEnv("PROCESSED_DIR", "data/processed"),
		ChunkSize:    getEnvInt("CHUNK_SIZE", 500),
		ChunkOverlap: getEnvInt("CHUNK_OVERLAP", 50),
	}
}

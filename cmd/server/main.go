
package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/qdrant/go-client/qdrant"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/api"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/config"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/embedder"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/llm"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/retriever"
)

func main() {
	cfg := config.Load()

	addr := flag.String("addr", cfg.ServerAddress, "listen address")
	qdrantHost := flag.String("qdrant-host", cfg.QdrantHost, "qdrant host")
	qdrantPort := flag.Int("qdrant-port", cfg.QdrantPort, "qdrant grpc port")
	topK := flag.Uint64("k", cfg.TopK, "chunks to retrieve per query")
	flag.Parse()

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: *qdrantHost,
		Port: *qdrantPort,
	})
	if err != nil {
		log.Fatal(err)
	}

	emb := embedder.NewOllama(cfg.OllamaURL, cfg.EmbeddingModel)
	ret := retriever.New(client, cfg.CollectionName)
	gen := llm.NewOllama(cfg.OllamaURL, cfg.LLMModel)

	server := api.New(emb, ret, gen, *topK)

	slog.Info("starting server", "addr", *addr)

	if err := http.ListenAndServe(*addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

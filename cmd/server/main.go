package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"

	"github.com/qdrant/go-client/qdrant"
	"github.com/vishalpatel08/rag-service-go/internal/api"
	"github.com/vishalpatel08/rag-service-go/internal/embedder"
	"github.com/vishalpatel08/rag-service-go/internal/llm"
	"github.com/vishalpatel08/rag-service-go/internal/retriever"
)

const collectionName = "distsys-corpus"

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	qdrantHost := flag.String("qdrant-host", "localhost", "qdrant host")
	qdrantPort := flag.Int("qdrant-port", 6334, "qdrant grpc port")
	topK := flag.Uint64("k", 5, "chunks to retrieve per query")
	flag.Parse()

	client, err := qdrant.NewClient(&qdrant.Config{Host: *qdrantHost, Port: *qdrantPort})
	if err != nil {
		log.Fatal(err)
	}

	emb := embedder.NewOllama("nomic-embed-text")
	ret := retriever.New(client, collectionName)
	gen := llm.NewOllama("llama3.2")

	server := api.New(emb, ret, gen, *topK)

	slog.Info("starting server", "addr", *addr)
	if err := http.ListenAndServe(*addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

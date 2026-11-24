
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/qdrant/go-client/qdrant"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/config"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/embedder"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/llm"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/retriever"
)

func main() {
	cfg := config.Load()

	question := flag.String("q", "", "question to ask")
	topK := flag.Uint64("k", cfg.TopK, "number of chunks to retrieve")
	flag.Parse()

	if *question == "" {
		log.Fatal("pass a question with -q")
	}

	ctx := context.Background()

	// Embed the question
	emb := embedder.NewOllama(cfg.OllamaURL, cfg.EmbeddingModel)
	vectors, err := emb.Embed([]string{*question})
	if err != nil {
		log.Fatal(err)
	}
	queryVector := vectors[0]

	// Connect to Qdrant
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.QdrantHost,
		Port: cfg.QdrantPort,
	})
	if err != nil {
		log.Fatal(err)
	}

	r := retriever.New(client, cfg.CollectionName)

	results, err := r.Search(ctx, queryVector, *topK)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		fmt.Println("no relevant chunks found")
		return
	}

	// Build context from retrieved chunks
	var contextBuilder strings.Builder
	for i, res := range results {
		fmt.Fprintf(
			&contextBuilder,
			"[%d] source: %s\n%s\n\n",
			i+1,
			res.Source,
			res.Text,
		)
	}

	prompt := fmt.Sprintf(`Answer the question using only the context below. If the context doesn't contain the answer, say so.

Context:
%s
Question: %s`,
		contextBuilder.String(),
		*question,
	)

	// Generate the answer
	gen := llm.NewOllama(cfg.OllamaURL, cfg.LLMModel)
	answer, err := gen.Generate(prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Answer:")
	fmt.Println(answer)

	fmt.Println("\nSources:")
	for _, res := range results {
		fmt.Printf("- %s (chunk %d, score %.3f)\n",
			res.Source, res.Position, res.Score)
	}
}

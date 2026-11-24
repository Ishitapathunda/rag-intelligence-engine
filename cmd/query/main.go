package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/qdrant/go-client/qdrant"
	"github.com/vishalpatel08/rag-service-go/internal/embedder"
	"github.com/vishalpatel08/rag-service-go/internal/llm"
	"github.com/vishalpatel08/rag-service-go/internal/retriever"
)

const collectionName = "distsys-corpus"

func main() {
	question := flag.String("q", "", "question to ask")
	topK := flag.Uint64("k", 5, "number of chunks to retrieve")
	flag.Parse()

	if *question == "" {
		log.Fatal("pass a question with -q")
	}

	ctx := context.Background()

	// embed the question with the same model used during ingestion
	emb := embedder.NewOllama("nomic-embed-text")
	vectors, err := emb.Embed([]string{*question})
	if err != nil {
		log.Fatal(err)
	}
	queryVector := vectors[0]

	// search qdrant for the closest chunks
	client, err := qdrant.NewClient(&qdrant.Config{Host: "localhost", Port: 6334})
	if err != nil {
		log.Fatal(err)
	}
	r := retriever.New(client, collectionName)

	results, err := r.Search(ctx, queryVector, *topK)
	if err != nil {
		log.Fatal(err)
	}

	if len(results) == 0 {
		fmt.Println("no relevant chunks found")
		return
	}

	// build the prompt: context chunks + question
	var contextBuilder strings.Builder
	for i, res := range results {
		fmt.Fprintf(&contextBuilder, "[%d] source: %s\n%s\n\n", i+1, res.Source, res.Text)
	}

	prompt := fmt.Sprintf(`Answer the question using only the context below. If the context doesn't contain the answer, say so.

Context:
%s
Question: %s`, contextBuilder.String(), *question)

	// generate the answer
	gen := llm.NewOllama("llama3.2")
	answer, err := gen.Generate(prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Answer:")
	fmt.Println(answer)

	fmt.Println("\nSources:")
	for _, res := range results {
		fmt.Printf("- %s (chunk %d, score %.3f)\n", res.Source, res.Position, res.Score)
	}
}

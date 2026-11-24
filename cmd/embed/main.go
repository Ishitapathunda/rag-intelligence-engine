package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/chunker"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/embedder"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/config"
)

type embeddedChunk struct {
	chunker.Chunk
	Vector []float32 `json:"vector"`
}

func main() {
	cfg := config.Load()
	inFile := flag.String("in", "data/processed/chunks.json", "chunks.json from cmd/chunk")
	outFile := flag.String("out", "data/processed/embeddings.json", "where to write embedded chunks")
	model := flag.String("model", cfg.EmbeddingModel, "ollama embedding model")
	batchSize := flag.Int("batch", 20, "chunks per call")
	workers := flag.Int("workers", 4, "concurrent batches in flight")
	
	flag.Parse()

	data, err := os.ReadFile(*inFile)
	if err != nil {
		log.Fatal(err)
	}

	var chunks []chunker.Chunk
	if err := json.Unmarshal(data, &chunks); err != nil {
		log.Fatal(err)
	}

	emb := embedder.NewOllama(cfg.OllamaURL, *model)
	// split into batches
	var batches [][]chunker.Chunk
	for i := 0; i < len(chunks); i += *batchSize {
		end := i + *batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batches = append(batches, chunks[i:end])
	}

	results := make([]embeddedChunk, len(chunks))

	var wg sync.WaitGroup
	sem := make(chan struct{}, *workers)

	for batchIdx, batch := range batches {
		wg.Add(1)
		sem <- struct{}{}

		go func(batchIdx int, batch []chunker.Chunk) {
			defer wg.Done()
			defer func() { <-sem }()

			texts := make([]string, len(batch))
			for i, c := range batch {
				texts[i] = c.Text
			}

			vectors, err := emb.Embed(texts)
			if err != nil {
				log.Printf("batch %d failed: %v", batchIdx, err)
				return
			}

			for i, c := range batch {
				idx := batchIdx*(*batchSize) + i
				results[idx] = embeddedChunk{Chunk: c, Vector: vectors[i]}
			}

			fmt.Printf("batch %d/%d done (%d chunks)\n", batchIdx+1, len(batches), len(batch))
		}(batchIdx, batch)
	}

	wg.Wait()

	f, err := os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(results); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("wrote %d embedded chunks to %s\n", len(results), *outFile)
}

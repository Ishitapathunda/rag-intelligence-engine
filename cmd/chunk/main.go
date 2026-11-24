package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/chunker"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/config"
)

func main() {
	cfg := config.Load()

	inDir := flag.String("in", cfg.ProcessedDir, "directory containing clean .txt files")
	outFile := flag.String("out", filepath.Join(cfg.ProcessedDir, "chunks.json"), "output chunks JSON file")
	size := flag.Int("size", cfg.ChunkSize, "words per chunk")
	overlap := flag.Int("overlap", cfg.ChunkOverlap, "word overlap between chunks")
	flag.Parse()

	if *size <= 0 || *overlap < 0 || *overlap >= *size {
		log.Fatal("invalid chunk settings: size must be positive and overlap must be >= 0 and smaller than size")
	}

	entries, err := os.ReadDir(*inDir)
	if err != nil {
		log.Fatal(err)
	}

	var allChunks []chunker.Chunk

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".txt") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(*inDir, e.Name()))
		if err != nil {
			log.Printf("skipping %s: %v", e.Name(), err)
			continue
		}

		source := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		chunks := chunker.ChunkText(source, string(data), *size, *overlap)

		fmt.Printf("%s -> %d chunks\n", e.Name(), len(chunks))
		allChunks = append(allChunks, chunks...)
	}

	if err := os.MkdirAll(filepath.Dir(*outFile), 0755); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(allChunks); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("wrote %d chunks to %s\n", len(allChunks), *outFile)
}

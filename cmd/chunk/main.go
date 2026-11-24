package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vishalpatel08/rag-service-go/internal/chunker"
)

func main() {
	inDir := flag.String("in", "data/processed", "dir with clean .txt files")
	outFile := flag.String("out", "data/processed/chunks.json", "where to write chunks")
	size := flag.Int("size", 500, "words per chunk")
	overlap := flag.Int("overlap", 50, "word overlap between chunks")
	flag.Parse()

	entries, err := os.ReadDir(*inDir)
	if err != nil {
		log.Fatal(err)
	}

	var allChunks []chunker.Chunk

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(*inDir, e.Name()))
		if err != nil {
			log.Printf("skipping %s: %v", e.Name(), err)
			continue
		}

		source := strings.TrimSuffix(e.Name(), ".txt")
		chunks := chunker.ChunkText(source, string(data), *size, *overlap)
		fmt.Printf("%s -> %d chunks\n", e.Name(), len(chunks))
		allChunks = append(allChunks, chunks...)
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


package main

import (
	"context"
	"encoding/json"
	"flag"
	"hash/fnv"
	"log"
	"os"

	"github.com/qdrant/go-client/qdrant"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/config"
)

type embeddedChunk struct {
	ID       string    `json:"id"`
	Source   string    `json:"source"`
	Text     string    `json:"text"`
	Position int       `json:"position"`
	Vector   []float32 `json:"vector"`
}

const vectorSize = 768 // nomic-embed-text dimension

func main() {
	cfg := config.Load()

	inFile := flag.String("in", "data/processed/embeddings.json", "embeddings.json from cmd/embed")
	host := flag.String("host", cfg.QdrantHost, "qdrant host")
	port := flag.Int("port", cfg.QdrantPort, "qdrant grpc port")
	flag.Parse()

	data, err := os.ReadFile(*inFile)
	if err != nil {
		log.Fatal(err)
	}

	var chunks []embeddedChunk
	if err := json.Unmarshal(data, &chunks); err != nil {
		log.Fatal(err)
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: *host,
		Port: *port,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	exists, err := client.CollectionExists(ctx, cfg.CollectionName)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		err = client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: cfg.CollectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     vectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("created collection %q", cfg.CollectionName)
	}

	var points []*qdrant.PointStruct

	for _, c := range chunks {
		if len(c.Vector) == 0 {
			continue // batch failed during embedding, skip it
		}

		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(numericID(c.ID)),
			Vectors: qdrant.NewVectors(c.Vector...),
			Payload: qdrant.NewValueMap(map[string]any{
				"chunk_id": c.ID,
				"source":   c.Source,
				"text":     c.Text,
				"position": c.Position,
			}),
		})
	}

	_, err = client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: cfg.CollectionName,
		Points:         points,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("upserted %d points into %q", len(points), cfg.CollectionName)
}

func numericID(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

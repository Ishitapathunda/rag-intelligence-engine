package main

import (
	"context"
	"encoding/json"
	"flag"
	"hash/fnv"
	"log"
	"os"

	"github.com/qdrant/go-client/qdrant"
)

type embeddedChunk struct {
	ID       string    `json:"id"`
	Source   string    `json:"source"`
	Text     string    `json:"text"`
	Position int       `json:"position"`
	Vector   []float32 `json:"vector"`
}

const collectionName = "distsys-corpus"
const vectorSize = 768 // nomic-embed-text dimension

func main() {
	inFile := flag.String("in", "data/processed/embeddings.json", "embeddings.json from cmd/embed")
	host := flag.String("host", "localhost", "qdrant host")
	port := flag.Int("port", 6334, "qdrant grpc port")
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

	exists, err := client.CollectionExists(ctx, collectionName)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		err = client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: collectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     vectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("created collection %q", collectionName)
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
		CollectionName: collectionName,
		Points:         points,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("upserted %d points into %q", len(points), collectionName)
}

func numericID(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

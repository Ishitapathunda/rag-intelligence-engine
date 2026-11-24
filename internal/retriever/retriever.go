package retriever

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
)

type Result struct {
	Text     string
	Source   string
	Position int
	Score    float32
}

type Retriever struct {
	client         *qdrant.Client
	collectionName string
}

func New(client *qdrant.Client, collectionName string) *Retriever {
	return &Retriever{client: client, collectionName: collectionName}
}

// Search returns the topK chunks closest to queryVector.
func (r *Retriever) Search(ctx context.Context, queryVector []float32, topK uint64) ([]Result, error) {
	points, err := r.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: r.collectionName,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          &topK,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(points))
	for _, p := range points {
		payload := p.GetPayload()
		results = append(results, Result{
			Text:     payload["text"].GetStringValue(),
			Source:   payload["source"].GetStringValue(),
			Position: int(payload["position"].GetIntegerValue()),
			Score:    p.GetScore(),
		})
	}

	return results, nil
}

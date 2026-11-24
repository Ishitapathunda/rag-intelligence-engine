package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/vishalpatel08/rag-service-go/internal/embedder"
	"github.com/vishalpatel08/rag-service-go/internal/llm"
	"github.com/vishalpatel08/rag-service-go/internal/retriever"
)

type Server struct {
	embedder  *embedder.OllamaEmbedder
	retriever *retriever.Retriever
	llm       *llm.OllamaLLM
	topK      uint64
	logger    *slog.Logger
}

func New(emb *embedder.OllamaEmbedder, ret *retriever.Retriever, gen *llm.OllamaLLM, topK uint64) *Server {
	return &Server{
		embedder:  emb,
		retriever: ret,
		llm:       gen,
		topK:      topK,
		logger:    slog.Default(),
	}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /query", s.handleQuery)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

type queryRequest struct {
	Question string `json:"question"`
}

type queryResponse struct {
	Answer  string   `json:"answer"`
	Sources []source `json:"sources"`
}

type source struct {
	Source   string  `json:"source"`
	Position int     `json:"position"`
	Score    float32 `json:"score"`
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
		http.Error(w, "invalid request: expected {\"question\": \"...\"}", http.StatusBadRequest)
		return
	}

	start := time.Now()

	// embed the question
	t0 := time.Now()
	vectors, err := s.embedder.Embed([]string{req.Question})
	if err != nil {
		s.logger.Error("embed failed", "err", err)
		http.Error(w, "failed to embed question", http.StatusInternalServerError)
		return
	}
	embedMs := time.Since(t0).Milliseconds()

	// retrieve relevant chunks
	t0 = time.Now()
	results, err := s.retriever.Search(ctx, vectors[0], s.topK)
	if err != nil {
		s.logger.Error("retrieval failed", "err", err)
		http.Error(w, "failed to search index", http.StatusInternalServerError)
		return
	}
	searchMs := time.Since(t0).Milliseconds()

	if len(results) == 0 {
		json.NewEncoder(w).Encode(queryResponse{Answer: "I don't have relevant information to answer that.", Sources: nil})
		return
	}

	// build prompt and generate
	var contextBuilder strings.Builder
	for i, res := range results {
		fmt.Fprintf(&contextBuilder, "[%d] source: %s\n%s\n\n", i+1, res.Source, res.Text)
	}
	prompt := fmt.Sprintf(`Answer the question using only the context below. If the context doesn't contain the answer, say so.

Context:
%s
Question: %s`, contextBuilder.String(), req.Question)

	t0 = time.Now()
	answer, err := s.llm.Generate(prompt)
	if err != nil {
		s.logger.Error("generation failed", "err", err)
		http.Error(w, "failed to generate answer", http.StatusInternalServerError)
		return
	}
	generateMs := time.Since(t0).Milliseconds()

	sources := make([]source, len(results))
	for i, res := range results {
		sources[i] = source{Source: res.Source, Position: res.Position, Score: res.Score}
	}

	s.logger.Info("query handled",
		"embed_ms", embedMs,
		"search_ms", searchMs,
		"generate_ms", generateMs,
		"total_ms", time.Since(start).Milliseconds(),
	)

	json.NewEncoder(w).Encode(queryResponse{Answer: answer, Sources: sources})
}

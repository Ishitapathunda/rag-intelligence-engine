package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/embedder"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/llm"
	"github.com/Ishitapathunda/rag-intelligence-engine/internal/retriever"
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

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req queryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeJSON(w, http.StatusRequestEntityTooLarge,
				errorResponse{Error: "request body is too large"})
			return
		}

		writeJSON(w, http.StatusBadRequest,
			errorResponse{Error: "invalid JSON request"})
		return
	}

	// Reject multiple JSON values in one request.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest,
			errorResponse{Error: "request must contain only one JSON object"})
		return
	}

	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		writeJSON(w, http.StatusBadRequest,
			errorResponse{Error: "question cannot be empty"})
		return
	}

	start := time.Now()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Step 1: Generate question embedding.
	t0 := time.Now()
	vectors, err := s.embedder.Embed([]string{req.Question})
	if err != nil {
		s.logger.Error("embedding failed", "error", err)
		writeJSON(w, http.StatusBadGateway,
			errorResponse{Error: "failed to generate question embedding"})
		return
	}
	embedMs := time.Since(t0).Milliseconds()

	if len(vectors) == 0 || len(vectors[0]) == 0 {
		writeJSON(w, http.StatusBadGateway,
			errorResponse{Error: "embedding service returned an empty vector"})
		return
	}

	// Step 2: Retrieve relevant document chunks.
	t0 = time.Now()
	results, err := s.retriever.Search(ctx, vectors[0], s.topK)
	if err != nil {
		s.logger.Error("retrieval failed", "error", err)

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			writeJSON(w, http.StatusGatewayTimeout,
				errorResponse{Error: "retrieval timed out"})
			return
		}

		writeJSON(w, http.StatusBadGateway,
			errorResponse{Error: "failed to search document index"})
		return
	}
	searchMs := time.Since(t0).Milliseconds()

	// Step 3: Handle queries with no relevant documents.
	if len(results) == 0 {
		s.logger.Info("no relevant documents found",
			"total_ms", time.Since(start).Milliseconds())

		writeJSON(w, http.StatusOK, queryResponse{
			Answer:  "I don't have relevant information to answer that.",
			Sources: []source{},
		})
		return
	}

	// Step 4: Build context for the LLM.
	var contextBuilder strings.Builder

	for i, res := range results {
		fmt.Fprintf(&contextBuilder,
			"[%d] source: %s\n%s\n\n",
			i+1, res.Source, res.Text)
	}

	prompt := fmt.Sprintf(`Answer the question using only the context below.
If the context does not contain the answer, clearly say so.
Do not invent facts.

Context:
%s

Question: %s`, contextBuilder.String(), req.Question)

	// Step 5: Generate the final answer.
	t0 = time.Now()
	answer, err := s.llm.Generate(prompt)
	if err != nil {
		s.logger.Error("LLM generation failed", "error", err)
		writeJSON(w, http.StatusBadGateway,
			errorResponse{Error: "failed to generate answer"})
		return
	}
	generateMs := time.Since(t0).Milliseconds()

	// Step 6: Prepare source citations.
	sources := make([]source, len(results))

	for i, res := range results {
		sources[i] = source{
			Source:   res.Source,
			Position: res.Position,
			Score:    res.Score,
		}
	}

	// Step 7: Log performance metrics.
	totalMs := time.Since(start).Milliseconds()

	s.logger.Info("query handled successfully",
		"embed_ms", embedMs,
		"search_ms", searchMs,
		"generate_ms", generateMs,
		"total_ms", totalMs,
		"source_count", len(sources),
	)

	writeJSON(w, http.StatusOK, queryResponse{
		Answer:  answer,
		Sources: sources,
	})
}

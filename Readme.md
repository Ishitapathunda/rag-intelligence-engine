# distsys-rag

A RAG (Retrieval-Augmented Generation) service that answers questions over a corpus of
distributed systems papers (Raft, Spanner, Dynamo, Bigtable, ...) and real production incident
postmortems (AWS, Cloudflare, GitHub). Built as a Go backend service, not a Python notebook —
ingestion, embedding orchestration, retrieval, and the API layer are all Go.

## Why this corpus

Most "chat with PDF" RAG demos use generic documents where you can't tell a good answer from a
hallucinated one. This corpus is deliberately technical and verifiable: every answer can be
checked against a specific paper section or incident report, which makes it possible to build a
real evaluation set (ground-truth Q&A pairs) rather than eyeballing outputs.

## Architecture

```
raw docs (pdf/txt/md)
      │  cmd/ingest — pdftotext extraction, cleanup
      ▼
clean text (.txt)
      │  cmd/chunk — fixed-size chunking w/ overlap
      ▼
chunks.json
      │  cmd/embed — Ollama (nomic-embed-text), concurrent batches
      ▼
embeddings.json
      │  cmd/store — Qdrant collection create + upsert
      ▼
Qdrant (vector DB)
      │  cmd/server — HTTP API
      │    embed query → search Qdrant → build prompt → generate (Ollama llama3.1)
      ▼
POST /query → { answer, sources }
```

## Stack

- **Language:** Go
- **Vector DB:** Qdrant
- **Embeddings + generation:** Ollama (local, `nomic-embed-text` + `llama3.1`) — no external API
  cost
- **API:** plain `net/http`, Go 1.22+ method-based routing

## Running it

```bash
# 1. start Qdrant
docker-compose up -d

# 2. pull local models
ollama pull nomic-embed-text
ollama pull llama3.1

# 3. run the pipeline once (raw docs -> Qdrant)
go run ./cmd/ingest
go run ./cmd/chunk
go run ./cmd/embed
go run ./cmd/store

# 4. start the API
go run ./cmd/server
```

Query it:
```bash
curl -X POST localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"question": "How does Raft handle leader election?"}'
```

## Project layout

```
cmd/
  ingest/   raw pdf/txt/md -> clean .txt
  chunk/    clean .txt -> chunks.json
  embed/    chunks.json -> embeddings.json
  store/    embeddings.json -> Qdrant
  query/    CLI for manual testing
  server/   HTTP API
internal/
  chunker/    chunking logic
  embedder/   Ollama embedding client
  llm/        Ollama generation client
  retriever/  Qdrant search wrapper
  api/        HTTP handlers
```

## Notes on design choices

- **Chunking:** fixed-size word windows with overlap, not semantic chunking - kept simple for v1,
  a natural next iteration.
- **Concurrency:** embedding generation batches requests and runs them concurrently via a
  worker-pool pattern (semaphore + WaitGroup), since embedding APIs are I/O-bound and benefit from
  parallel calls.
- **Idempotent ingestion:** chunk IDs are deterministic (hash of source + position), so re-running
  ingestion overwrites existing Qdrant points instead of duplicating them.
- **Local-first:** embeddings and generation both run through Ollama, so the whole pipeline works
  offline and at zero API cost.

## Planned next

- Evaluation harness (RAGAS) against a ground-truth Q&A set built from the corpus
- Request tracing (Langfuse) across the embed/search/generate stages
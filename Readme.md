# RAG Intelligence Engine

A Go-based Retrieval-Augmented Generation (RAG) application that enables users to ask questions over documents and receive context-aware answers using locally hosted Large Language Models.

The system combines document processing, text chunking, embedding generation, vector search, and LLM-based response generation into a modular backend pipeline.

## Overview

RAG Intelligence Engine uses a retrieval-based approach to provide answers grounded in information extracted from documents.

Instead of relying only on an LLM's pretrained knowledge, the application retrieves relevant document chunks from a vector database and uses them as context when generating responses.

The backend is implemented in Go, with Ollama handling embedding generation and language model inference, and Qdrant managing vector storage and similarity search.

## Key Features

- **Document Ingestion:** Extract text from PDF, TXT, and Markdown files.
- **Text Chunking:** Split documents into smaller chunks using configurable chunk sizes and overlap.
- **Embedding Generation:** Generate vector embeddings using Ollama.
- **Concurrent Processing:** Process embedding batches concurrently using goroutines, WaitGroups, and semaphores.
- **Vector Storage:** Store and retrieve document embeddings using Qdrant.
- **Semantic Retrieval:** Retrieve the most relevant document chunks for a user query.
- **Context-Aware Generation:** Generate answers using retrieved document context.
- **Source References:** Return source document names, chunk positions, and relevance scores.
- **HTTP API:** Expose document-question answering functionality through REST endpoints.
- **Centralized Configuration:** Configure models, database connection, chunking parameters, and server settings using environment variables.

## Architecture

```text
              DOCUMENT INPUT
             PDF / TXT / MD
                    |
                    v
             Document Ingestion
               cmd/ingest
                    |
                    v
              Clean Text Files
                    |
                    v
                Chunking
                cmd/chunk
                    |
                    v
                chunks.json
                    |
                    v
          Embedding Generation
                cmd/embed
             (Ollama + Go)
                    |
                    v
              embeddings.json
                    |
                    v
              Vector Storage
                cmd/store
                    |
                    v
              QDRANT DATABASE
                    |
                    |
USER QUESTION
      |
      v
Query Embedding
      |
      v
Semantic Retrieval
      |
      v
Relevant Document Chunks
      |
      v
Context Construction
      |
      v
LLM Response Generation
      |
      v
Answer + Source References
```

## Technology Stack

| Component | Technology |
|---|---|
| Backend Language | Go |
| Vector Database | Qdrant |
| Embedding Model | nomic-embed-text |
| Language Model | llama3.1 |
| Local Model Runtime | Ollama |
| HTTP Server | Go net/http |
| Containerization | Docker |
| Configuration | Environment Variables |

## Project Structure

```text
rag-intelligence-engine/
|
├── cmd/
│   ├── ingest/       # Document ingestion
│   ├── chunk/        # Text chunking
│   ├── embed/        # Concurrent embedding generation
│   ├── store/        # Qdrant storage
│   ├── query/        # CLI-based querying
│   └── server/       # HTTP API server
|
├── internal/
│   ├── api/          # HTTP handlers and API tests
│   ├── chunker/      # Text chunking logic
│   ├── config/       # Centralized application configuration
│   ├── embedder/     # Ollama embedding client
│   ├── llm/          # Ollama generation client
│   └── retriever/    # Qdrant retrieval logic
|
├── data/
│   ├── raw/          # Input documents
│   └── processed/    # Processed text and embeddings
|
├── DockerFile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Prerequisites

Install the following before running the application:

- Go
- Docker and Docker Compose
- Ollama
- Qdrant

Pull the required Ollama models:

```bash
ollama pull nomic-embed-text
ollama pull llama3.1
```

## Installation and Setup

Clone the repository:

```bash
git clone https://github.com/Ishitapathunda/rag-intelligence-engine.git

cd rag-intelligence-engine
```

Download dependencies:

```bash
go mod download
```

Start Qdrant:

```bash
docker compose up -d
```

Make sure Ollama is running locally.

## Running the Application

Execute the processing pipeline in sequence.

**1. Document ingestion**

Place your PDF, TXT, or Markdown documents inside `data/raw/`.

```bash
go run ./cmd/ingest
```

**2. Generate text chunks**

```bash
go run ./cmd/chunk
```

**3. Generate embeddings**

```bash
go run ./cmd/embed
```

**4. Store embeddings in Qdrant**

```bash
go run ./cmd/store
```

**5. Start the API server**

```bash
go run ./cmd/server
```

The server runs on port `8080` by default.

## API Documentation

### Health Check

**Endpoint:** `GET /health`

Example:

```bash
curl http://localhost:8080/health
```

### Ask a Question

**Endpoint:** `POST /query`

Request:

```json
{
  "question": "What is the main challenge in distributed systems?"
}
```

Example:

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"question":"What is the main challenge in distributed systems?"}'
```

Response structure:

```json
{
  "answer": "Generated answer based on retrieved document context.",
  "sources": [
    {
      "source": "document-name",
      "position": 1,
      "score": 0.85
    }
  ]
}
```

## Configuration

The application supports environment-based configuration.

| Variable | Default |
|---|---|
| OLLAMA_URL | http://localhost:11434 |
| EMBEDDING_MODEL | nomic-embed-text |
| LLM_MODEL | llama3.1 |
| QDRANT_HOST | localhost |
| QDRANT_PORT | 6334 |
| COLLECTION_NAME | distsys-corpus |
| TOP_K | 5 |
| SERVER_ADDRESS | :8080 |
| RAW_DIR | data/raw |
| PROCESSED_DIR | data/processed |
| CHUNK_SIZE | 500 |
| CHUNK_OVERLAP | 50 |

## Running Tests

Execute all Go tests:

```bash
go test ./...
```

Build the complete project:

```bash
go build ./...
```

## Design Considerations

- **Modular Architecture:** Separate packages handle ingestion, chunking, embeddings, retrieval, and API operations.
- **Concurrent Embedding:** Goroutines and synchronization primitives allow multiple embedding batches to run concurrently.
- **Deterministic Chunk IDs:** Chunk identifiers are generated using document source and chunk position.
- **Local Model Execution:** Ollama provides local embedding and text-generation capabilities.
- **Configurable Pipeline:** Application settings can be adjusted through environment variables.

## Future Improvements

- Retrieval and generation evaluation using a ground-truth question-answer dataset.
- Improved document retrieval and ranking.
- Request tracing across the RAG pipeline.

## Author

**Ishita Pathunda**

GitHub: [@Ishitapathunda](https://github.com/Ishitapathunda)

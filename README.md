# ollama-mem

A lightweight Go CLI that gives Ollama persistent memory — RAG over local embeddings, no vector DB required.

## Prerequisites

```bash
ollama pull nomic-embed-text
ollama pull qwen2.5-coder:7b
```

Make sure Ollama is running (`ollama serve` or the background app).

## Build

```bash
go build -o ollama-mem .
```

## Usage

```bash
# Store a memory
./ollama-mem remember "I'm working on secure-deploy-kit, an IaC project with LocalStack"
./ollama-mem remember "I prefer Go for personal projects, Java/Spring Boot for the local job market"

# Ask a question with automatic context injection
./ollama-mem ask "What project am I working on right now?"

# List all stored memories
./ollama-mem list

# Wipe everything
./ollama-mem clear
```

Memories are stored in `~/.ollama-mem/memory.jsonl` (JSON Lines, one entry per line with text, embedding vector, and timestamp).

## How it works

1. `remember` sends the text to `nomic-embed-text` via `/api/embed`, then stores the vector and text in a local file.
2. `ask` embeds your question, computes cosine similarity against every stored memory, keeps the top 5 (score > 0.3), and injects them into the system prompt sent to `qwen2.5-coder:7b` via `/api/chat`.
3. No external dependencies — just the Go stdlib and Ollama's local HTTP API.

## Known limitations (MVP)

- Linear search (O(n) per query) — plenty fast for a few thousand entries, but not built to scale indefinitely. For more: index with a real vector DB (Qdrant, Chroma) or a local HNSW.
- No deduplication or summarization — memories accumulate as-is.
- `qwen2.5-coder:7b` is only used for chat, not embeddings (it isn't built for that).

## Roadmap

- Interactive `chat` REPL mode (conversation history + memory in the same loop)
- Bubbletea TUI frontend (consistent with `tunepipe`)
- Periodic summarization of old memories to avoid raw accumulation
- Tags/categories on memories for filtered retrieval
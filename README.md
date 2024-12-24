# GhastlyDB- a super lightweight key-value vector DB

![build](https://github.com/aakashshankar/vexdb/actions/workflows/build_and_test.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/aakashshankar/vexdb/badge.svg?branch=tests)](https://coveralls.io/github/aakashshankar/vexdb?branch=tests)


A Go-based vector database and search engine that supports multiple embedding providers and efficient vector similarity search.

## Features 💪

- Multiple embedding providers supported:
  - OpenAI
  - NVIDIA
- In-memory and persistent storage options
- SSTable-based storage for durability
- Cosine similarity search
- Thread-safe operations with concurrent access support
- BERT tokenization support

## Architecture 💻

The system consists of several key components:

- **Embedding Layer**: Supports multiple embedding providers (OpenAI, NVIDIA, ColBERT)
- **Storage Layer**: 
  - SSTable-based persistent storage
  - Memtable for write buffering
- **Search**: Implements cosine similarity for vector comparison
- **Tokenization**: BERT-based tokenization support

## Installation 💾

### Prerequisites

- Go 1.21 or higher
- Make
- golangci-lint (for development)

### From Source

1. Clone the repository:
```bash
git clone https://github.com/ahhcash/ghastly.git 
cd ghastly
```

2. Build the binary:
```bash
make build
```

This will create a binary in the `bin/` directory for your current OS and architecture.

### Cross-Platform Builds

To build for multiple platforms:

```bash
make build-all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64) 
- Windows (amd64)

### Development Setup 

1. Install dependencies: 
```bash
make deps
```

2. Run tests:
```bash
make test
```

3. Run linting and formatting:
```bash
make lint
make fmt
```

## Usage 🧑‍💻

`/path/to/repo/bin/ghastly <your command> [flags] <your input>`

Add `/path/to/repo/bin` your `$PATH` for convenience

## Storage Design 💽

The storage system uses a LSM-tree inspired design:
1. Writes go to an in-memory Memtable
2. When Memtable is full, it's flushed to disk as an SSTable
3. Background compaction process merges SSTables

## Contributing 🙏

lmao

Just clone the repo and you're good to "go" 😏

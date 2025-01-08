# GhastlyDB - a super lightweight vector database in Go

![build](https://github.com/ahhcash/ghastly/actions/workflows/build_and_deploy.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/ahhcash/ghastly/badge.svg?branch=master)](https://coveralls.io/github/ahhcash/ghastly?branch=master)

I've built this as an experiment - to truly understand how databases work. This is only possible if I built it from first principles.
GhastlyDB is the result of this experiment, and I'm super excited about how it turned out.

## Features 💪

### Embedding Support
- Multiple embedding providers:
  - OpenAI (using text-embedding-3-small model)
  - NVIDIA (using nv-embedqa-mistral-7b-v2)
  - ColBERT (local embedding support)

### Storage Engine
- LSM Tree-based storage architecture
- Memory-mapped memtable for fast writes
- SSTable-based persistent storage
- Skip list implementation for efficient data structure
- Thread-safe operations with concurrent access support

### Search Capabilities
- Multiple similarity metrics:
  - Cosine similarity
  - Dot product
  - L2 distance
- Efficient vector comparison algorithms
- Sorted search results with similarity scores

### Cross-Platform Support
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Installation 💾

### Prerequisites

- Go 1.21 or higher
- Make
- pkg-config

### Local inference specific dependencies
- ONNX Runtime (for local embedding model inference)
- Make sure `libtokenizers.a` is present inside `/libs/static/libotkenizers`. You can build it from [source](https://github.com/daulet/tokenizers)
  or find it in the [releases](https://github.com/daulet/tokenizers/releases/) page of HuggingFace's tokenizers port for Go. (shoutout @daulet)

### Platform-Specific Dependencies

#### macOS
```bash
brew install pkg-config
brew install onnxruntime
```

#### Linux
```bash
sudo apt-get update
sudo apt-get install build-essential pkg-config
pip install onnxruntime
```

#### Windows
```batch
pip install onnxruntime
```

### Building From Source

1. Clone the repository:
```bash
git clone https://github.com/ahhcash/ghastly.git
cd ghastly
```

2. Build for your platform:
```bash
make build
```

This will create a binary in the `bin/` directory for your current OS and architecture.

3. Build for all platforms:
```bash
make build-all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Usage 🧑‍💻
GhastlyDB provides a REPL interface for interactive use:
```bash
./bin/ghastly
```
Available commands:

* `put <key> <value>` - Store a key-value pair
* `get <key>` - Retrieve a value by key
* `search <query>` - Perform semantic search
* `delete <key>` - Deletes a key
* `help` - Provides a list of valid commands
* `exit` - Exit the REPL


## Configuration
Default configuration:

```go
Config{
Path:           "./ghastlydb_data",
MemtableSize:   64 * 1024 * 1024, // 64MB
Metric:         "cosine",
EmbeddingModel: "openai",
}
```

## API Usage (Coming soon 🤫)
```go
import "github.com/ahhcash/ghastlydb/db"

// Initialize with default config
database, err := db.OpenDB(db.DefaultConfig())

// Store data
err = database.Put("key", "value")

// Retrieve data
value, err := database.Get("key")

// Semantic search
results, err := database.Search("query")
```
## Architecture 🛠️
### Storage Layer
GhastlyDB uses a Log-Structured Merge Tree (LSM) architecture:

Writes are buffered in an in-memory memtable (implemented as a skip list)
When memtable reaches its size limit, it's flushed to disk as an SSTable
SSTables are immutable and contain sorted key-value pairs
Background processes handle SSTable compaction

### Search Engine
The search implementation supports multiple distance metrics:

Cosine similarity for normalized vectors
Dot product for raw similarity
L2 distance for Euclidean space

### Embedding Layer

**OpenAI**: Cloud-based embeddings using text-embedding-3-small <br>
**NVIDIA**: Cloud-based embeddings using nv-embedqa-mistral-7b-v2 <br>
**ColBERT**: Local inference using ONNX runtime, libtokenizers on colBERT-ir/v2

## Development
### Testing
```bash
make test        # Run tests
make coverage    # Generate coverage report
```

### Code Quality
```bash
make lint        # Run golangci-lint
make fmt         # Format code
```

## Directory Structure
```
Directory structure:
└── ahhcash-ghastly/
├── README.md
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── .golangci.yml
├── clients/
│   └── python/
│       ├── __init__.py
│       ├── client.py
│       ├── setup.py
│       └── test_client.py
├── cmd/
│   └── main.go
├── db/
│   ├── db.go
│   └── db_test.go
├── embed/
│   ├── embedder.go
│   ├── local/
│   │   └── colbert/
│   │       ├── config.go
│   │       ├── darwin.go
│   │       ├── embed.go
│   │       ├── linux.go
│   │       ├── platform_specific.go
│   │       └── windows.go
│   ├── nvidia/
│   │   ├── embed.go
│   │   └── types.go
│   └── openai/
│       ├── embed.go
│       └── types.go
├── grpc/
│   ├── gen/
│   │   └── grpc/
│   │       └── proto/
│   │           ├── ghastly.pb.go
│   │           └── ghastly_grpc.pb.go
│   ├── proto/
│   │   └── ghastly.proto
│   └── server/
│       └── server.go
├── http/
│   └── server/
│       └── server.go
├── index/
│   ├── connections.go
│   ├── hnsw.go
│   └── search.go
├── libs/
│   └── static/
│       └── libtokenizers/
│           └── .gitkeep
├── mocks/
│   └── embedder.go
├── search/
│   ├── cosine.go
│   ├── dot.go
│   ├── l2.go
│   └── metrics_test.go
├── storage/
│   ├── memtable.go
│   ├── memtable_test.go
│   ├── skiplist.go
│   ├── skiplist_test.go
│   ├── sstable.go
│   ├── store.go
│   └── store_test.go
└── .github/
└── workflows/
└── build_and_deploy.yml

```
## Contributing 🙏

I would absolutely love any feedback / contributions! Please open a PR, and I'll gladly take a look :)
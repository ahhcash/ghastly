package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/aakashshankar/vexdb/embed/nvidia"
	"github.com/aakashshankar/vexdb/search"
	"github.com/spf13/cobra"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var searchCmd = &cobra.Command{
	Use:   "search [embeddings file] [query]",
	Short: "Perform a semantic search based in the query",
	Long:  `This command performs a semantic search based on the query.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		embeddingsDir := args[0]
		query := args[1]

		embedder, err := nvidia.LoadNvidiaEmbedder()
		if err != nil {
			return fmt.Errorf("error initializing NVIDIA embedder: %w", err)
		}

		queryEmbedding, err := embedder.Embed(query)
		if err != nil {
			return fmt.Errorf("error generating embeddings: %w", err)
		}

		files, err := os.ReadDir(embeddingsDir)

		var best string
		var maxScore float64

		for _, file := range files {
			embeddings, err := readEmbeddingsFile(filepath.Join(embeddingsDir, file.Name()))

			if err != nil {
				fmt.Printf("error reading file: %v", file.Name())
				continue
			}

			for path, emb := range embeddings {
				score := search.Cosine(queryEmbedding, emb)
				if score > maxScore {
					best = path
					maxScore = score
				}
			}
		}

		if best == "" {
			fmt.Printf("no match found for query %s\n", query)
		} else {
			fmt.Printf("%s -> %f\n", best, maxScore)
		}

		return nil
	},
}

func readEmbeddingsFile(filePath string) (map[string][]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening embeddings file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	scanner := bufio.NewScanner(file)
	embeddings := make(map[string][]float64)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " : ", 2)
		if len(parts) != 2 {
			continue
		}

		path := parts[0]
		embeddingBytes, err := base64.StdEncoding.DecodeString(parts[1])

		if err != nil {
			return nil, err
		}

		embedding := make([]float64, len(embeddingBytes)/8)
		for i := 0; i < len(embedding); i++ {
			embedding[i] = math.Float64frombits(binary.LittleEndian.Uint64(embeddingBytes[i*8 : (i+1)*8]))
		}

		embeddings[path] = embedding
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading embeddings file: %w", err)
	}

	return embeddings, nil
}

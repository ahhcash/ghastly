package cmd

import (
	"fmt"
	"github.com/aakashshankar/vexdb/pkg/embed/nvidia"
	"github.com/aakashshankar/vexdb/pkg/search"
	"github.com/spf13/cobra"
	"os"
)

var searchCmd = &cobra.Command{
	Use:   "search [file path] [query]",
	Short: "Perform a semantic search based in the query",
	Long:  `This command performs a semantic search based on the query.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		query := args[1]

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		embedder, err := nvidia.LoadNvidiaEmbedder()
		if err != nil {
			return fmt.Errorf("error initializing NVIDIA embedder: %w", err)
		}

		queryEmbedding, err := embedder.Embed([]string{query})
		if err != nil {
			return fmt.Errorf("error generating embeddings: %w", err)
		}

		embeddings, err := embedder.Embed([]string{string(content)})
		if err != nil {
			return fmt.Errorf("error generating embeddings: %w", err)
		}

		similarity := search.Cosine(queryEmbedding, embeddings)
		fmt.Printf("Similarity score: %f\n", similarity)
		return nil
	},
}

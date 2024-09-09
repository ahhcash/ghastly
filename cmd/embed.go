package cmd

import (
	"fmt"
	"github.com/aakashshankar/vexdb/pkg/embed/nvidia"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var embedCmd = &cobra.Command{
	Use:   "embed [file path]",
	Short: "Generate embeddings for a file",
	Long:  `This command generates embeddings for the content of the specified file using the NVIDIA embedder.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}
		// Initialize NVIDIA embedder
		embedder, err := nvidia.LoadNvidiaEmbedder()
		if err != nil {
			return fmt.Errorf("error initializing NVIDIA embedder: %w", err)
		}

		// Generate embeddings
		embeddings, err := embedder.Embed([]string{string(content)})
		if err != nil {
			return fmt.Errorf("error generating embeddings: %w", err)
		}

		// Print embeddings
		fmt.Printf("Embeddings for %s:\n", filePath)
		fmt.Println(strings.Trim(fmt.Sprint(embeddings), "[]"))

		return nil
	},
}

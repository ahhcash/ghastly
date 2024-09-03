package cmd

import (
	"fmt"
	"github.com/aakashshankar/vexdb/pkg/embed/nvidia"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "vexdb",
	Short: "VexDB CLI tool",
	Long:  `A CLI tool for VexDB operations including embedding generation`,
}

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
		fmt.Println(string(content))
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(embedCmd)
}

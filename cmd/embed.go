package cmd

import (
	"encoding/binary"
	"fmt"
	"github.com/aakashshankar/vexdb/pkg/embed/nvidia"
	"github.com/aakashshankar/vexdb/pkg/storage"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
	"github.com/spf13/cobra"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var embedCmd = &cobra.Command{
	Use:   "embed [source path] [destination directory]",
	Short: "Generate embeddings for a file",
	Long:  `This command generates embeddings for the content of the specified file using the NVIDIA embedder.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceDir := args[0]
		destPath := args[1]

		files, err := os.ReadDir(destPath)

		if len(files) > 0 {
			return fmt.Errorf("destination directory not empty: %v", destPath)
		}

		_ = os.MkdirAll(destPath, 0777)

		embedder, err := nvidia.LoadNvidiaEmbedder()
		if err != nil {
			return fmt.Errorf("error initializing NVIDIA embedder: %w", err)
		}

		memtable := storage.NewMemtable(1000)

		err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))

				var content string

				switch ext {
				case ".txt", ".md", ".go", ".py", ".java", ".js", ".c", ".cpp", ".rs":
					data, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("could not read file %s: %v", path, err)
					}
					content = string(data)
				case ".pdf":
					content, err = extractTextFromPDF(path)
					if err != nil {
						return fmt.Errorf("could not extract text from pdf file %s: %v", path, err)
					}
				}

				embeddings, err := embedder.Embed(content)

				if err != nil {
					return fmt.Errorf("could not generate embeddings: %v", err)
				}

				key := path
				value := make([]byte, len(embeddings)*8)
				for i, v := range embeddings {
					binary.LittleEndian.PutUint64(value[i*8:], math.Float64bits(v))
				}
				memtable.Put(key, value, destPath)

				fmt.Printf("Generated and stored embeddings for %s\n", path)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("error processing files in %s: %w", sourceDir, err)
		}

		if memtable.Size() > 0 {
			err := memtable.FlushToDisk(filepath.Join(destPath, uuid.New().String()+".pkl"))
			if err != nil {
				return fmt.Errorf("could not flush memtable to disk: %v", err)
			}

		}

		fmt.Printf("Embeddings in %s stored at %s\n", sourceDir, destPath)

		return nil
	},
}

func extractTextFromPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	var builder strings.Builder
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(&builder, b)
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}

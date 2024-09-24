package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "vexdb",
	Short: "VexDB CLI tool",
	Long:  `A CLI tool for VexDB operations including embedding generation`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(searchCmd)
}

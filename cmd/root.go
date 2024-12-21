package cmd

import (
	"bufio"
	"fmt"
	db2 "github.com/aakashshankar/vexdb/db"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "vexdb",
	Short: "VexDB REPL CLI tool",
	Long:  `A CLI tool for VexDB operations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := repl()
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func repl() error {
	reader := bufio.NewReader(os.Stdin)
	db, err := db2.OpenDB(db2.DefaultConfig())
	if err != nil {
		return fmt.Errorf("error initializing database: %v", err)
	}
	fmt.Println("VexDB REPL. 'help' gives you command list")
	for {
		fmt.Print("vexdb> ")
		op, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %v", err)
		}

		input := strings.TrimSpace(op)
		if input == "exit" {
			break
		}

		if len(input) > 0 {
			err = processReplCommand(input, db)
		}
		if err != nil {
			fmt.Printf("%v", err)
		}
	}

	return nil
}

func processReplCommand(input string, db *db2.DB) error {
	args := strings.SplitN(input, " ", 3)
	if len(args) == 0 {
		return nil
	}
	cmd := args[0]

	args = args[1:]
	switch cmd {
	case "help":
		fmt.Println("Available commands:")
		fmt.Println("  put <key> <value> - Store a key-value pair")
		fmt.Println("  get <key>         - Retrieve a value by key")
		fmt.Println("  search <value>    - Semantically search for values")
		fmt.Println("  exit              - Exit the REPL")
		return nil
	case "put":
		if len(args) != 2 {
			return fmt.Errorf("'put' requires exactly 2 arguments: key and value\n")
		}
		return db.Put(args[0], args[1])
	case "get":
		if len(args) != 1 {
			return fmt.Errorf("'get' requires exactly 1 argument: key\n")
		}

		value, err := db.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	case "search":
		if len(args) != 1 {
			return fmt.Errorf("'search' takes exactly 1 argument\n")
		}
		res, err := db.Search(args[0])
		if err != nil {
			return fmt.Errorf("error during search: %v", err)
		}
		fmt.Println("KEY\tVALUE\tSCORE")
		for _, r := range res {
			fmt.Printf("%s\t%s\t%.2f\n", r.Key, r.Value, r.Score)
		}
	default:
		return fmt.Errorf("unknown command: %s\n", cmd)
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {

}

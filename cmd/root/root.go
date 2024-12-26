package root

import (
	"bufio"
	"fmt"
	db2 "github.com/ahhcash/ghastlydb/db"
	"github.com/ahhcash/ghastlydb/storage"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"unicode/utf8"
)

var rootCmd = &cobra.Command{
	Use:   "ghastly",
	Short: "GhastlyDB REPL CLI tool",
	Long:  `A CLI tool for GhastlyDB operations`,
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
	fmt.Println("GhastlyDB REPL. 'help' gives you command list")
	for {
		fmt.Print("ghastly> ")
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

func formatSearchResults(results []storage.Result) string {
	if len(results) == 0 {
		return "No results found"
	}

	headers := []string{"KEY", "VALUE", "SCORE"}

	maxKeyLen := len(headers[0])
	maxValueLen := len(headers[1])
	maxScoreLen := len(headers[2])

	for _, r := range results {
		keyLen := utf8.RuneCountInString(r.Key)
		valueLen := utf8.RuneCountInString(r.Value)
		scoreLen := len(fmt.Sprintf("%.4f", r.Score))

		if keyLen > maxKeyLen {
			maxKeyLen = keyLen
		}
		if valueLen > maxValueLen {
			maxValueLen = valueLen
		}
		if scoreLen > maxScoreLen {
			maxScoreLen = scoreLen
		}
	}

	const padding = 3
	maxKeyLen += padding
	maxValueLen += padding

	formatStr := fmt.Sprintf("%%-%ds%%-%ds%%%ds\n", maxKeyLen, maxValueLen, maxScoreLen)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(formatStr, headers[0], headers[1], headers[2]))

	separatorLen := maxKeyLen + maxValueLen + maxScoreLen
	sb.WriteString(strings.Repeat("-", separatorLen) + "\n")

	for _, r := range results {
		sb.WriteString(fmt.Sprintf(formatStr, r.Key, r.Value, fmt.Sprintf("%.4f", r.Score)))
	}

	return sb.String()
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
		fmt.Print(formatSearchResults(res))
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

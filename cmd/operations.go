package cmd

import (
	"fmt"
	"github.com/aakashshankar/vexdb/db"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put [key] [value]",
	Short: "Store a key-value pair in the database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := db.DefaultConfig()
		database, err := db.OpenDB(config)
		if err != nil {
			return fmt.Errorf("error opening database at %s: %w")
		}
		key := args[0]
		value := args[1]
		err = database.Put(key, value)

		if err != nil {
			return fmt.Errorf("error on Put(%s, %s): %v", key, value, err)
		}

		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Fetch the value of a key from the database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := db.DefaultConfig()
		database, err := db.OpenDB(config)
		if err != nil {
			return fmt.Errorf("error opening database: %v")
		}

		key := args[0]

		value, err := database.Get(key)
		if err != nil {
			return fmt.Errorf("error on Get(%s): %v", key, err)
		}

		fmt.Printf("Key: %s; Value: %s", key, value)

		return nil
	},
}

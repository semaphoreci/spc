package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func fetchRequiredStringFlag(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)

	if err != nil || value == "" {
		fmt.Printf("(err) %s path not provided\n", name)
		os.Exit(1)
	}

	return value
}

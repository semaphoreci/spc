package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// revive:disable:deep-exit
func fetchRequiredStringFlag(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)

	if err != nil || value == "" {
		fmt.Printf("(err) %s path not provided\n", name)
		os.Exit(1)
	}

	return value
}

func fetchOptionalStringFlag(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)

	if err != nil {
		fmt.Printf("(err) unable to fetch argument %s\n", name)
		os.Exit(1)
	}

	return value
}

func fetchOptionalBoolFlag(cmd *cobra.Command, name string) bool {
	value, err := cmd.Flags().GetBool(name)

	if err != nil {
		fmt.Printf("(err) unable to fetch argument %s\n", name)
		os.Exit(1)
	}

	return value
}

// revive:enable:deep-exit

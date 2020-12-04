package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spc",
	Short: "Semaphore 2.0 Pipeline Compiler",
}

func Execute() error {
	return rootCmd.Execute()
}

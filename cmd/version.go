package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "0.1.3"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dashmin %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
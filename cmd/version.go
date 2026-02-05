package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.3.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dashmin %s\n", Version)
	},
}


package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version = "1.0.0"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pmp300 version %s\n", Version)
		fmt.Println("Diamond Rio PMP300 Management Tool")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

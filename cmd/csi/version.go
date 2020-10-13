package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var short bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version information",
	Long: "print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start to print version info")
	},
}

func init() {
	cmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&short, "short", false, "print just the version number")
}
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gejie",
	Short: "scrape sites from command line using Golang; eCommerce product data, job listings, etc.",
	Long:  "scrape sites from command line using Golang; eCommerce product data, job listings, etc. add more details later",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

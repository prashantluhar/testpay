package cli

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "testpay",
	Short: "TestPay — mock payment gateway for local development and CI",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(scenarioCmd)
	rootCmd.AddCommand(logsCmd)
}

package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var scenarioCmd = &cobra.Command{
	Use:   "scenario",
	Short: "Manage and run scenarios",
}

var scenarioListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scenarios",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Calls GET /api/scenarios on the running server
		return callAPI("GET", "http://localhost:7700/api/scenarios", nil)
	},
}

var scenarioRunCmd = &cobra.Command{
	Use:   "run [scenario-id]",
	Short: "Run a scenario",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := fmt.Sprintf("http://localhost:7700/api/scenarios/%s/run", args[0])
		return callAPI("POST", url, nil)
	},
}

func init() {
	scenarioCmd.AddCommand(scenarioListCmd)
	scenarioCmd.AddCommand(scenarioRunCmd)
}

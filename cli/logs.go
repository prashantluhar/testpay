package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var followLogs bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View request logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if followLogs {
			return tailLogs()
		}
		return callAPI("GET", "http://localhost:7700/api/logs", nil)
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Tail live logs")
}

func tailLogs() error {
	seen := map[string]bool{}
	for {
		resp, err := http.Get("http://localhost:7700/api/logs?limit=20")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		var logs []map[string]any
		json.NewDecoder(resp.Body).Decode(&logs)
		resp.Body.Close()
		for _, l := range logs {
			id, _ := l["id"].(string)
			if !seen[id] {
				seen[id] = true
				fmt.Printf("%s %s %s %v\n", l["method"], l["path"], l["response_status"], l["duration_ms"])
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func callAPI(method, url string, body io.Reader) error {
	req, _ := http.NewRequest(method, url, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling %s: %w", url, err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
	fmt.Println()
	return nil
}

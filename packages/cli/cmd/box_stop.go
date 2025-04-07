package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/cli/config"
	"github.com/spf13/cobra"
)

type BoxStopOptions struct {
	OutputFormat string
}

func NewBoxStopCommand() *cobra.Command {
	opts := &BoxStopOptions{}

	cmd := &cobra.Command{
		Use:   "stop [box-id]",
		Short: "Stop a running box",
		Long:  "Stop a running box by its ID",
		Example: `  gbox box stop 550e8400-e29b-41d4-a716-446655440000
  gbox box stop 550e8400-e29b-41d4-a716-446655440000 --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(args[0], opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.OutputFormat, "output", "text", "Output format (json or text)")

	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runStop(boxID string, opts *BoxStopOptions) error {
	apiBase := config.GetAPIURL()
	apiURL := fmt.Sprintf("%s/api/v1/boxes/%s/stop", strings.TrimSuffix(apiBase, "/"), boxID)

	if os.Getenv("DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Request URL: %s\n", apiURL)
	}

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("API call failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Response status code: %d\n", resp.StatusCode)
		fmt.Fprintf(os.Stderr, "Response content: %s\n", string(body))
	}

	return handleStopResponse(resp.StatusCode, body, boxID, opts.OutputFormat)
}

func handleStopResponse(statusCode int, body []byte, boxID, outputFormat string) error {
	switch statusCode {
	case 200:
		if outputFormat == "json" {
			fmt.Println(`{"status":"success","message":"Box stopped successfully"}`)
		} else {
			fmt.Println("Box stopped successfully")
		}
	case 404:
		fmt.Printf("Box not found: %s\n", boxID)
	default:
		errorMsg := fmt.Sprintf("Error: Failed to stop box (HTTP %d)", statusCode)
		if os.Getenv("DEBUG") == "true" {
			errorMsg = fmt.Sprintf("%s\nResponse: %s", errorMsg, string(body))
		}
		fmt.Println(errorMsg)
	}

	return nil
}

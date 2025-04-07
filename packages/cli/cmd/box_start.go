package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/cli/config"
	"github.com/spf13/cobra"
)

type BoxStartOptions struct {
	OutputFormat string
}

type BoxStartResponse struct {
	Message string `json:"message"`
}

func NewBoxStartCommand() *cobra.Command {
	opts := &BoxStartOptions{}

	cmd := &cobra.Command{
		Use:   "start [box-id]",
		Short: "Start a stopped box",
		Long:  "Start a stopped box by its ID",
		Example: `  gbox box start 550e8400-e29b-41d4-a716-446655440000
  gbox box start 550e8400-e29b-41d4-a716-446655440000 --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(args[0], opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.OutputFormat, "output", "text", "Output format (json or text)")

	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runStart(boxID string, opts *BoxStartOptions) error {
	apiBase := config.GetAPIURL()
	apiURL := fmt.Sprintf("%s/api/v1/boxes/%s/start", strings.TrimSuffix(apiBase, "/"), boxID)

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

	return handleStartResponse(resp.StatusCode, body, boxID, opts.OutputFormat)
}

func handleStartResponse(statusCode int, body []byte, boxID, outputFormat string) error {
	switch statusCode {
	case 200:
		if outputFormat == "json" {
			fmt.Println(string(body))
		} else {
			var response BoxStartResponse
			if err := json.Unmarshal(body, &response); err != nil {
				fmt.Println("Box started successfully")
			} else {
				fmt.Println(response.Message)
			}
		}
	case 404:
		fmt.Printf("Box not found: %s\n", boxID)
	case 400:
		if strings.Contains(string(body), "already running") {
			fmt.Printf("Box is already running: %s\n", boxID)
		} else {
			fmt.Printf("Error: Invalid request: %s\n", string(body))
		}
	default:
		errorMsg := fmt.Sprintf("Error: Failed to start box (HTTP %d)", statusCode)
		if os.Getenv("DEBUG") == "true" {
			errorMsg = fmt.Sprintf("%s\nResponse: %s", errorMsg, string(body))
		}
		fmt.Println(errorMsg)
	}

	return nil
}

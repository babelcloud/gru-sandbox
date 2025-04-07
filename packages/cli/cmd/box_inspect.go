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

type BoxInspectOptions struct {
	OutputFormat string
}

func NewBoxInspectCommand() *cobra.Command {
	opts := &BoxInspectOptions{}

	cmd := &cobra.Command{
		Use:   "inspect [box-id]",
		Short: "Get detailed information about a box",
		Long:  "Get detailed information about a box by its ID",
		Example: `  gbox box inspect 550e8400-e29b-41d4-a716-446655440000              # Get box details
  gbox box inspect 550e8400-e29b-41d4-a716-446655440000 --output json  # Get box details in JSON format`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(args[0], opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.OutputFormat, "output", "text", "Output format (json or text)")

	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runInspect(boxID string, opts *BoxInspectOptions) error {
	apiBase := config.GetAPIURL()
	apiURL := fmt.Sprintf("%s/api/v1/boxes/%s", strings.TrimSuffix(apiBase, "/"), boxID)

	if os.Getenv("DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Request URL: %s\n", apiURL)
	}

	resp, err := http.Get(apiURL)
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

	return handleInspectResponse(resp.StatusCode, body, boxID, opts.OutputFormat)
}

func handleInspectResponse(statusCode int, body []byte, boxID, outputFormat string) error {
	switch statusCode {
	case 200:
		if outputFormat == "json" {
			// Output JSON directly
			fmt.Println(string(body))
		} else {
			// Output in text format
			fmt.Println("Box details:")
			fmt.Println("------------")

			// Parse JSON and format output
			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return fmt.Errorf("failed to parse JSON response: %v", err)
			}

			// Output each key-value pair
			for key, value := range data {
				// Handle complex types
				var valueStr string
				switch v := value.(type) {
				case string, float64, bool, int:
					valueStr = fmt.Sprintf("%v", v)
				default:
					// For objects or arrays, use JSON format
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						valueStr = fmt.Sprintf("%v", v)
					} else {
						valueStr = string(jsonBytes)
					}
				}
				fmt.Printf("%-15s: %s\n", key, valueStr)
			}
		}
	case 404:
		fmt.Printf("Box not found: %s\n", boxID)
	default:
		errorMsg := fmt.Sprintf("Error: Failed to get box details (HTTP %d)", statusCode)
		if os.Getenv("DEBUG") == "true" {
			errorMsg = fmt.Sprintf("%s\nResponse: %s", errorMsg, string(body))
		}
		fmt.Println(errorMsg)
	}

	return nil
}

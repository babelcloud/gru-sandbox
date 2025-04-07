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

type BoxReclaimOptions struct {
	OutputFormat string
	Force        bool
}

type BoxReclaimResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	StoppedCount int    `json:"stoppedCount"`
	DeletedCount int    `json:"deletedCount"`
}

func NewBoxReclaimCommand() *cobra.Command {
	opts := &BoxReclaimOptions{}

	cmd := &cobra.Command{
		Use:   "reclaim [box-id]",
		Short: "Reclaim a box resources",
		Long:  "Reclaim a box's resources by force if it's in a stuck state",
		Example: `  gbox box reclaim 550e8400-e29b-41d4-a716-446655440000              # Reclaim box resources
  gbox box reclaim 550e8400-e29b-41d4-a716-446655440000 --force      # Force reclaim box resources
  gbox box reclaim 550e8400-e29b-41d4-a716-446655440000 --output json  # Output result in JSON format
  gbox box reclaim                                      # Reclaim resources for all eligible boxes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boxID := ""
			if len(args) > 0 {
				boxID = args[0]
			}
			return runReclaim(boxID, opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.OutputFormat, "output", "text", "Output format (json or text)")
	flags.BoolVarP(&opts.Force, "force", "f", false, "Force resource reclamation, even if box is running")

	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runReclaim(boxID string, opts *BoxReclaimOptions) error {
	apiURL := buildReclaimAPIURL(boxID, opts.Force)

	if os.Getenv("DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Request URL: %s\n", apiURL)
	}

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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

	return handleReclaimResponse(resp.StatusCode, body, boxID, opts.OutputFormat)
}

func buildReclaimAPIURL(boxID string, force bool) string {
	apiBase := config.GetAPIURL()
	var apiURL string

	if boxID == "" {
		// If no box ID specified, perform global reclaim
		apiURL = fmt.Sprintf("%s/api/v1/boxes/reclaim", strings.TrimSuffix(apiBase, "/"))
	} else {
		// If box ID specified, reclaim only that specific box
		apiURL = fmt.Sprintf("%s/api/v1/boxes/%s/reclaim", strings.TrimSuffix(apiBase, "/"), boxID)
	}

	// Add force parameter
	if force {
		if strings.Contains(apiURL, "?") {
			apiURL += "&force=true"
		} else {
			apiURL += "?force=true"
		}
	}

	return apiURL
}

func handleReclaimResponse(statusCode int, body []byte, boxID, outputFormat string) error {
	switch statusCode {
	case 200:
		if outputFormat == "json" {
			// Output JSON directly
			fmt.Println(string(body))
		} else {
			// Output in text format
			var response BoxReclaimResponse
			if err := json.Unmarshal(body, &response); err != nil {
				fmt.Println("Box resources successfully reclaimed")
			} else {
				fmt.Println(response.Message)
				if response.StoppedCount > 0 {
					fmt.Printf("Stopped %d boxes\n", response.StoppedCount)
				}
				if response.DeletedCount > 0 {
					fmt.Printf("Deleted %d boxes\n", response.DeletedCount)
				}
			}
		}
	case 404:
		if boxID != "" {
			fmt.Printf("Box not found: %s\n", boxID)
		} else {
			fmt.Println("No boxes found to reclaim")
		}
	case 400:
		fmt.Printf("Error: Invalid request: %s\n", string(body))
	default:
		errorMsg := fmt.Sprintf("Error: Failed to reclaim box resources (HTTP %d)", statusCode)
		if os.Getenv("DEBUG") == "true" {
			errorMsg = fmt.Sprintf("%s\nResponse: %s", errorMsg, string(body))
		}
		fmt.Println(errorMsg)
	}

	return nil
}

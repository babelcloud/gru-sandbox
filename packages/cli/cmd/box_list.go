package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type BoxResponse struct {
	Boxes []struct {
		ID     string `json:"id"`
		Image  string `json:"image"`
		Status string `json:"status"`
	} `json:"boxes"`
}

func NewBoxListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "list",
		Short:              "List all available boxes",
		Long:               "List all available boxes with various filtering options",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string
			var filters []string

			// Parse arguments, supporting the same format as in bash script
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--output":
					if i+1 < len(args) {
						outputFormat = args[i+1]
						if outputFormat != "json" && outputFormat != "text" {
							fmt.Println("Error: Invalid output format. Must be 'json' or 'text'")
							os.Exit(1)
						}
						i++
					} else {
						fmt.Println("Error: --output requires a value")
						os.Exit(1)
					}
				case "-f", "--filter":
					if i+1 < len(args) {
						filter := args[i+1]
						if !strings.Contains(filter, "=") {
							fmt.Println("Error: Invalid filter format. Use field=value")
							os.Exit(1)
						}
						filters = append(filters, filter)
						i++
					} else {
						fmt.Println("Error: --filter requires a value")
						os.Exit(1)
					}
				case "--help":
					printBoxListHelp()
					return
				default:
					fmt.Printf("Error: Unknown option %s\n", args[i])
					os.Exit(1)
				}
			}

			// If output format not specified, default to text
			if outputFormat == "" {
				outputFormat = "text"
			}

			// Build query parameters
			queryParams := ""
			if len(filters) > 0 {
				for i, f := range filters {
					parts := strings.SplitN(f, "=", 2)
					if len(parts) == 2 {
						field := parts[0]
						value := url.QueryEscape(parts[1])

						if i == 0 {
							queryParams = "?filter=" + field + "=" + value
						} else {
							queryParams += "&filter=" + field + "=" + value
						}
					}
				}
			}

			// Call API server
			apiURL := "http://localhost:28080/api/v1/boxes" + queryParams
			if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
				apiURL = envURL + "/api/v1/boxes" + queryParams
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "Request URL: %s\n", apiURL)
			}

			resp, err := http.Get(apiURL)
			if err != nil {
				fmt.Printf("Error: API call failed: %v\n", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error: Failed to read response: %v\n", err)
				os.Exit(1)
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "Response status code: %d\n", resp.StatusCode)
				fmt.Fprintf(os.Stderr, "Response content: %s\n", string(body))
			}

			// Handle HTTP status code
			switch resp.StatusCode {
			case 200:
				var response BoxResponse
				if err := json.Unmarshal(body, &response); err != nil {
					fmt.Printf("Error: Failed to parse JSON response: %v\n", err)
					os.Exit(1)
				}

				if outputFormat == "json" {
					// JSON format output
					fmt.Println(string(body))
				} else {
					// Text format output
					if len(response.Boxes) == 0 {
						fmt.Println("No boxes found")
						return
					}

					// Print header
					fmt.Println("ID                                      IMAGE               STATUS")
					fmt.Println("---------------------------------------- ------------------- ---------------")

					// Print each box's information
					for _, box := range response.Boxes {
						image := box.Image
						if strings.HasPrefix(image, "sha256:") {
							image = strings.TrimPrefix(image, "sha256:")
							if len(image) > 12 {
								image = image[:12]
							}
						}
						fmt.Printf("%-40s %-20s %s\n", box.ID, image, box.Status)
					}
				}
			case 404:
				fmt.Println("No boxes found")
			default:
				fmt.Printf("Error: Failed to get box list (HTTP %d)\n", resp.StatusCode)
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
				}
				os.Exit(1)
			}
		},
	}

	return cmd
}

func printBoxListHelp() {
	fmt.Println("Usage: gbox box list [options]")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("    gbox box list                              # List all boxes")
	fmt.Println("    gbox box list --output json                # List boxes in JSON format")
	fmt.Println("    gbox box list -f 'label=project=myapp'     # List boxes with label project=myapp")
	fmt.Println("    gbox box list -f 'ancestor=ubuntu:latest'  # List boxes using ubuntu:latest image")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("    --output          Output format (json or text, default: text)")
	fmt.Println("    -f, --filter      Filter boxes (format: field=value)")
	fmt.Println("                      Supported fields: id, label, ancestor")
	fmt.Println("                      Examples:")
	fmt.Println("                      -f 'id=abc123'")
	fmt.Println("                      -f 'label=project=myapp'")
	fmt.Println("                      -f 'ancestor=ubuntu:latest'")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    --help            Show help information")
}

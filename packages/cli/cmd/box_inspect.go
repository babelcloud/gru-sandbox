package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewBoxInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "inspect",
		Short:              "Get detailed information about a box",
		Long:               "Get detailed information about a box by its ID",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string

			// Parse arguments
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxInspectHelp()
					return
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
				default:
					if !strings.HasPrefix(args[i], "-") && boxID == "" {
						boxID = args[i]
					} else if strings.HasPrefix(args[i], "-") {
						fmt.Printf("Error: Unknown option %s\n", args[i])
						os.Exit(1)
					} else {
						fmt.Printf("Error: Unexpected argument %s\n", args[i])
						os.Exit(1)
					}
				}
			}

			// Validate box ID
			if boxID == "" {
				fmt.Println("Error: Box ID is required")
				os.Exit(1)
			}

			// Call API to get box details
			apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s", boxID)
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = fmt.Sprintf("%s/api/v1/boxes/%s", envURL, boxID)
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
						fmt.Printf("Error: Failed to parse JSON response: %v\n", err)
						os.Exit(1)
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
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			default:
				fmt.Printf("Error: Failed to get box details (HTTP %d)\n", resp.StatusCode)
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
				}
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}
		},
	}

	return cmd
}

func printBoxInspectHelp() {
	fmt.Println("Usage: gbox box inspect <id> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    --output          Output format (json or text, default: text)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("    gbox box inspect 550e8400-e29b-41d4-a716-446655440000              # Get box details")
	fmt.Println("    gbox box inspect 550e8400-e29b-41d4-a716-446655440000 --output json  # Get box details in JSON format")
	fmt.Println()
}

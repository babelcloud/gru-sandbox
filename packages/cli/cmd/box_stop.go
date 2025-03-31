package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewBoxStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "stop",
		Short:              "Stop a running box",
		Long:               "Stop a running box by its ID",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string

			// Parse arguments
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxStopHelp()
					return
				case "--output":
					if i+1 < len(args) {
						outputFormat = args[i+1]
						if outputFormat != "json" && outputFormat != "text" {
							fmt.Println("Error: Invalid output format. Must be 'json' or 'text'")
							if os.Getenv("TESTING") != "true" {
								os.Exit(1)
							}
							return
						}
						i++
					} else {
						fmt.Println("Error: --output requires a value")
						if os.Getenv("TESTING") != "true" {
							os.Exit(1)
						}
						return
					}
				default:
					if !strings.HasPrefix(args[i], "-") && boxID == "" {
						boxID = args[i]
					} else if strings.HasPrefix(args[i], "-") {
						fmt.Printf("Error: Unknown option %s\n", args[i])
						if os.Getenv("TESTING") != "true" {
							os.Exit(1)
						}
						return
					} else {
						fmt.Printf("Error: Unexpected argument %s\n", args[i])
						if os.Getenv("TESTING") != "true" {
							os.Exit(1)
						}
						return
					}
				}
			}

			// Validate box ID
			if boxID == "" {
				fmt.Println("Error: Box ID is required")
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}

			// Call API to stop the box
			apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s/stop", boxID)
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = fmt.Sprintf("%s/api/v1/boxes/%s/stop", envURL, boxID)
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "Request URL: %s\n", apiURL)
			}

			// Create POST request
			req, err := http.NewRequest("POST", apiURL, nil)
			if err != nil {
				fmt.Printf("Error: Failed to create request: %v\n", err)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}
			req.Header.Set("Content-Type", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error: API call failed: %v\n", err)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error: Failed to read response: %v\n", err)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "Response status code: %d\n", resp.StatusCode)
				fmt.Fprintf(os.Stderr, "Response content: %s\n", string(body))
			}

			// Handle HTTP status code
			switch resp.StatusCode {
			case 200:
				if outputFormat == "json" {
					fmt.Println(`{"status":"success","message":"Box stopped successfully"}`)
				} else {
					fmt.Println("Box stopped successfully")
				}
			case 404:
				fmt.Printf("Box not found: %s\n", boxID)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			default:
				fmt.Printf("Error: Failed to stop box (HTTP %d)\n", resp.StatusCode)
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

func printBoxStopHelp() {
	fmt.Println("Usage: gbox box stop <id> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    --output          Output format (json or text, default: text)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("    gbox box stop 550e8400-e29b-41d4-a716-446655440000              # Stop a box")
	fmt.Println("    gbox box stop 550e8400-e29b-41d4-a716-446655440000 --output json  # Stop a box and output JSON")
	fmt.Println()
}

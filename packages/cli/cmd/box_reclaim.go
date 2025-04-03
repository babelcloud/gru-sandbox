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

type BoxReclaimResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	StoppedCount int    `json:"stoppedCount"`
	DeletedCount int    `json:"deletedCount"`
}

func NewBoxReclaimCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "reclaim",
		Short:              "Reclaim a box resources",
		Long:               "Reclaim a box's resources by force if it's in a stuck state",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string
			var force bool = false

			// Parse arguments
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxReclaimHelp()
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
				case "--force", "-f":
					force = true
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

			// Prepare API URL
			var apiURL string
			if boxID == "" {
				// If no box ID specified, perform global reclaim
				apiURL = "http://localhost:28080/api/v1/boxes/reclaim"
				if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
					apiURL = fmt.Sprintf("%s/api/v1/boxes/reclaim", envURL)
				}
			} else {
				// If box ID specified, reclaim only that specific box
				apiURL = fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s/reclaim", boxID)
				if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
					apiURL = fmt.Sprintf("%s/api/v1/boxes/%s/reclaim", envURL, boxID)
				}
			}

			// Add force parameter
			if force {
				if strings.Contains(apiURL, "?") {
					apiURL += "&force=true"
				} else {
					apiURL += "?force=true"
				}
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
			req.Header.Set("Accept", "application/json")

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
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			case 400:
				fmt.Printf("Error: Invalid request: %s\n", string(body))
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			default:
				fmt.Printf("Error: Failed to reclaim box resources (HTTP %d)\n", resp.StatusCode)
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

func printBoxReclaimHelp() {
	fmt.Println("Usage: gbox box reclaim <id> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    --output          Output format (json or text, default: text)")
	fmt.Println("    -f, --force       Force resource reclamation, even if box is running")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("    gbox box reclaim 550e8400-e29b-41d4-a716-446655440000              # Reclaim box resources")
	fmt.Println("    gbox box reclaim 550e8400-e29b-41d4-a716-446655440000 --force      # Force reclaim box resources")
	fmt.Println("    gbox box reclaim 550e8400-e29b-41d4-a716-446655440000 --output json  # Output result in JSON format")
	fmt.Println("    gbox box reclaim                                      # Reclaim resources for all eligible boxes")
	fmt.Println()
}

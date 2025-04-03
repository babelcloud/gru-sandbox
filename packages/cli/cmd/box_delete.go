package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type BoxListResponse struct {
	Boxes []struct {
		ID string `json:"id"`
	} `json:"boxes"`
}

func NewBoxDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "delete",
		Short:              "Delete a box by its ID",
		Long:               "Delete a box by its ID or delete all boxes",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string
			var deleteAll bool
			var force bool

			// Parse arguments
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxDeleteHelp()
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
				case "--all":
					deleteAll = true
				case "--force":
					force = true
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

			// Validate arguments
			if deleteAll && boxID != "" {
				fmt.Println("Error: Cannot specify both --all and a box ID")
				os.Exit(1)
			}

			if !deleteAll && boxID == "" {
				fmt.Println("Error: Must specify either --all or a box ID")
				os.Exit(1)
			}

			// Handle deleting all boxes
			if deleteAll {
				// Get the list of all boxes
				apiURL := "http://localhost:28080/api/v1/boxes"
				if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
					apiURL = envURL + "/api/v1/boxes"
				}
				resp, err := http.Get(apiURL)
				if err != nil {
					fmt.Printf("Error: Failed to get box list: %v\n", err)
					os.Exit(1)
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error: Failed to read response: %v\n", err)
					os.Exit(1)
				}

				// Debug output
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "API response:\n")
					var prettyJSON bytes.Buffer
					if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
						fmt.Fprintln(os.Stderr, prettyJSON.String())
					} else {
						fmt.Fprintln(os.Stderr, string(body))
					}
				}

				var response BoxListResponse
				if err := json.Unmarshal(body, &response); err != nil {
					fmt.Printf("Error: Failed to parse JSON response: %v\n", err)
					os.Exit(1)
				}

				if len(response.Boxes) == 0 {
					if outputFormat == "json" {
						fmt.Println(`{"status":"success","message":"No boxes to delete"}`)
					} else {
						fmt.Println("No boxes to delete")
					}
					return
				}

				// Show boxes that will be deleted
				fmt.Println("The following boxes will be deleted:")
				for _, box := range response.Boxes {
					fmt.Printf("  - %s\n", box.ID)
				}
				fmt.Println()

				// If not forced, confirm deletion
				if !force {
					fmt.Print("Are you sure you want to delete all boxes? [y/N] ")
					reader := bufio.NewReader(os.Stdin)
					reply, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("Error: Failed to read input: %v\n", err)
						os.Exit(1)
					}

					reply = strings.TrimSpace(strings.ToLower(reply))
					if reply != "y" && reply != "yes" {
						if outputFormat == "json" {
							fmt.Println(`{"status":"cancelled","message":"Operation cancelled by user"}`)
						} else {
							fmt.Println("Operation cancelled")
						}
						return
					}
				}

				// Delete all boxes
				success := true
				for _, box := range response.Boxes {
					apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s", box.ID)
					if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
						apiURL = fmt.Sprintf("%s/api/v1/boxes/%s", envURL, box.ID)
					}
					req, err := http.NewRequest("DELETE", apiURL, strings.NewReader(`{"force":true}`))
					if err != nil {
						fmt.Printf("Error: Failed to create request: %v\n", err)
						success = false
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Printf("Error: Failed to delete box %s: %v\n", box.ID, err)
						success = false
						continue
					}
					resp.Body.Close()

					if resp.StatusCode != 200 && resp.StatusCode != 204 {
						fmt.Printf("Error: Failed to delete box %s, HTTP status code: %d\n", box.ID, resp.StatusCode)
						success = false
					}
				}

				if success {
					if outputFormat == "json" {
						fmt.Println(`{"status":"success","message":"All boxes deleted successfully"}`)
					} else {
						fmt.Println("All boxes deleted successfully")
					}
				} else {
					if outputFormat == "json" {
						fmt.Println(`{"status":"error","message":"Some boxes failed to delete"}`)
					} else {
						fmt.Println("Some boxes failed to delete")
					}
					os.Exit(1)
				}
				return
			}

			// Delete single box
			apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s", boxID)
			if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
				apiURL = fmt.Sprintf("%s/api/v1/boxes/%s", envURL, boxID)
			}
			req, err := http.NewRequest("DELETE", apiURL, strings.NewReader(`{"force":true}`))
			if err != nil {
				fmt.Printf("Error: Failed to create request: %v\n", err)
				os.Exit(1)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error: Failed to delete box. Make sure the API server is running and the ID '%s' is correct\n", boxID)
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 && resp.StatusCode != 204 {
				fmt.Printf("Error: Failed to delete box, HTTP status code: %d\n", resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				if os.Getenv("DEBUG") == "true" && len(body) > 0 {
					fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
				}
				os.Exit(1)
			}

			if outputFormat == "json" {
				fmt.Println(`{"status":"success","message":"Box deleted successfully"}`)
			} else {
				fmt.Println("Box deleted successfully")
			}
		},
	}

	return cmd
}

func printBoxDeleteHelp() {
	fmt.Println("Usage: gbox box delete [options] <id>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    --output          Output format (json or text, default: text)")
	fmt.Println("    --all             Delete all boxes")
	fmt.Println("    --force           Force deletion without confirmation")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("    gbox box delete 550e8400-e29b-41d4-a716-446655440000              # Delete a box")
	fmt.Println("    gbox box delete --all --force                                     # Delete all boxes without confirmation")
	fmt.Println("    gbox box delete --all                                             # Delete all boxes (requires confirmation)")
	fmt.Println("    gbox box delete 550e8400-e29b-41d4-a716-446655440000 --output json  # Delete a box and output JSON")
	fmt.Println()
}

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/api-server/models"

	"github.com/spf13/cobra"
)

type BoxCreateResponse struct {
	ID string `json:"id"`
}

func NewBoxCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "create",
		Short:              "Create a new box",
		Long:               "Create a new box with various options for image, environment, and commands",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string
			var image string
			var command string
			var commandArgs []string
			var env []string
			var labels []string
			var workingDir string

			// Parse arguments
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxCreateHelp()
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
				case "--image":
					if i+1 < len(args) {
						image = args[i+1]
						i++
					} else {
						fmt.Println("Error: --image requires a value")
						os.Exit(1)
					}
				case "--env":
					if i+1 < len(args) {
						envValue := args[i+1]
						if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.+$`).MatchString(envValue) {
							fmt.Println("Error: Invalid environment variable format. Use KEY=VALUE")
							os.Exit(1)
						}
						env = append(env, envValue)
						i++
					} else {
						fmt.Println("Error: --env requires a value")
						os.Exit(1)
					}
				case "-l", "--label":
					if i+1 < len(args) {
						labelValue := args[i+1]
						if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.+$`).MatchString(labelValue) {
							fmt.Println("Error: Invalid label format. Use KEY=VALUE")
							os.Exit(1)
						}
						labels = append(labels, labelValue)
						i++
					} else {
						fmt.Println("Error: --label requires a value")
						os.Exit(1)
					}
				case "-w", "--work-dir":
					if i+1 < len(args) {
						workingDir = args[i+1]
						i++
					} else {
						fmt.Println("Error: --work-dir requires a value")
						os.Exit(1)
					}
				case "--":
					if i+1 < len(args) {
						command = args[i+1]
						if i+2 < len(args) {
							commandArgs = args[i+2:]
						}
						i = len(args) // End the loop
					} else {
						fmt.Println("Error: -- requires a command")
						os.Exit(1)
					}
				default:
					fmt.Printf("Error: Unknown option %s\n", args[i])
					os.Exit(1)
				}
			}

			// If output format is not specified, default to text
			if outputFormat == "" {
				outputFormat = "text"
			}

			// Build request body
			request := models.BoxCreateRequest{}

			if image != "" {
				request.Image = image
			}

			if command != "" {
				request.Cmd = command
			}

			if len(commandArgs) > 0 {
				request.Args = commandArgs
			}

			if workingDir != "" {
				request.WorkingDir = workingDir
			}

			// 空值赋值给 ImagePullSecret，保证接口一致性
			request.ImagePullSecret = ""

			// Process environment variables
			if len(env) > 0 {
				request.Env = make(map[string]string)
				for _, e := range env {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						request.Env[parts[0]] = parts[1]
					}
				}
			}

			// Process labels
			if len(labels) > 0 {
				request.ExtraLabels = make(map[string]string)
				for _, l := range labels {
					parts := strings.SplitN(l, "=", 2)
					if len(parts) == 2 {
						request.ExtraLabels[parts[0]] = parts[1]
					}
				}
			}

			// Convert request to JSON
			requestBody, err := json.Marshal(request)
			if err != nil {
				fmt.Printf("Error: Unable to serialize request: %v\n", err)
				os.Exit(1)
			}

			// Debug output
			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "Request body:\n")
				var prettyJSON bytes.Buffer
				json.Indent(&prettyJSON, requestBody, "", "  ")
				fmt.Fprintln(os.Stderr, prettyJSON.String())
			}

			// Call API server
			apiURL := "http://localhost:28080/api/v1/boxes"
			if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
				apiURL = envURL + "/api/v1/boxes"
			}
			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				fmt.Printf("Error: Unable to connect to API server: %v\n", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			// Read response
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error: Failed to read response: %v\n", err)
				os.Exit(1)
			}

			// Check HTTP status code
			if resp.StatusCode != 201 {
				fmt.Printf("Error: API server returned HTTP %d\n", resp.StatusCode)
				if len(responseBody) > 0 {
					fmt.Printf("Response: %s\n", string(responseBody))
				}
				os.Exit(1)
			}

			// Process response
			if outputFormat == "json" {
				// Output JSON as is
				fmt.Println(string(responseBody))
			} else {
				// Format output
				var response BoxCreateResponse
				if err := json.Unmarshal(responseBody, &response); err != nil {
					fmt.Printf("Error: Failed to parse response: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Box created with ID \"%s\"\n", response.ID)
			}
		},
	}

	return cmd
}

func printBoxCreateHelp() {
	fmt.Println("Usage: gbox box create [options] [--] <command> [args...]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("    --output            Output format (json or text, default: text)")
	fmt.Println("    --image             Container image")
	fmt.Println("    --env               Environment variable, format KEY=VALUE")
	fmt.Println("    -w, --work-dir      Working directory")
	fmt.Println("    -l, --label         Custom label, format KEY=VALUE")
	fmt.Println("    --                  Command and its arguments")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("    gbox box create --image python:3.9 -- python3 -c 'print(\"Hello\")'")
	fmt.Println("    gbox box create --env PATH=/usr/local/bin:/usr/bin:/bin -w /app -- node server.js")
	fmt.Println("    gbox box create --label project=myapp --label env=prod -- python3 server.py")
	fmt.Println()
}

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/api-server/models"
	"github.com/babelcloud/gru-sandbox/packages/cli/config"
	"github.com/spf13/cobra"
)

type BoxCreateOptions struct {
	OutputFormat    string
	Image           string
	Env             []string
	Labels          []string
	WorkingDir      string
	Command         []string
	ImagePullSecret string
}

type BoxCreateResponse struct {
	ID string `json:"id"`
}

func parseKeyValuePairs(pairs []string, pairType string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}

	result := make(map[string]string)
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		} else {
			return nil, fmt.Errorf("invalid %s format: %s (must be KEY=VALUE)", pairType, pair)
		}
	}
	return result, nil
}

func NewBoxCreateCommand() *cobra.Command {
	opts := &BoxCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create [flags] -- [command] [args...]",
		Short: "Create a new box",
		Long: `Create a new box with various options for image, environment, and commands.

You can specify box configurations through various flags, including which container image to use,
setting environment variables, adding labels, and specifying a working directory.

Command arguments can be specified directly in the command line or added after the '--' separator.`,
		Example: `  gbox box create --image python:3.9 -- python3 -c 'print("Hello")'
  gbox box create --env PATH=/usr/local/bin:/usr/bin:/bin -w /app -- node server.js
  gbox box create --label project=myapp --label env=prod -- python3 server.py`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, args)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.OutputFormat, "output", "text", "Output format (json or text)")
	flags.StringVar(&opts.Image, "image", "", "Container image to use")
	flags.StringArrayVar(&opts.Env, "env", []string{}, "Environment variables in KEY=VALUE format")
	flags.StringArrayVarP(&opts.Labels, "label", "l", []string{}, "Custom labels in KEY=VALUE format")
	flags.StringVarP(&opts.WorkingDir, "work-dir", "w", "", "Working directory")

	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runCreate(opts *BoxCreateOptions, args []string) error {
	request := models.BoxCreateRequest{}

	request.Image = opts.Image
	request.ImagePullSecret = opts.ImagePullSecret

	envMap, err := parseKeyValuePairs(opts.Env, "environment variable")
	if err != nil {
		return err
	}
	request.Env = envMap

	labelMap, err := parseKeyValuePairs(opts.Labels, "label")
	if err != nil {
		return err
	}
	request.ExtraLabels = labelMap

	if opts.WorkingDir != "" {
		request.WorkingDir = opts.WorkingDir
	}

	if len(opts.Command) > 0 {
		request.Cmd = opts.Command[0]
		if len(opts.Command) > 1 {
			request.Args = opts.Command[1:]
		}
	} else if len(args) > 0 {
		request.Cmd = args[0]
		if len(args) > 1 {
			request.Args = args[1:]
		}
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("unable to serialize request: %v", err)
	}

	if os.Getenv("DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Request body:\n")
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, requestBody, "", "  ")
		fmt.Fprintln(os.Stderr, prettyJSON.String())
	}

	apiBase := config.GetAPIURL()
	apiURL := fmt.Sprintf("%s/api/v1/boxes", strings.TrimSuffix(apiBase, "/"))

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("unable to connect to API server: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 201 {
		errMsg := fmt.Sprintf("API server returned HTTP %d", resp.StatusCode)
		if len(responseBody) > 0 {
			errMsg += fmt.Sprintf("\nResponse: %s", string(responseBody))
		}
		return fmt.Errorf("%s", errMsg)
	}

	if opts.OutputFormat == "json" {
		fmt.Println(string(responseBody))
	} else {
		var response BoxCreateResponse
		if err := json.Unmarshal(responseBody, &response); err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}
		fmt.Printf("Box created with ID \"%s\"\n", response.ID)
	}
	return nil
}

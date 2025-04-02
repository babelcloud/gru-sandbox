package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// BoxPath represents the structure of a box path
type BoxPath struct {
	BoxID string
	Path  string
}

// Parse box path (format BOX_ID:PATH)
func parseBoxPath(path string) (*BoxPath, error) {
	re := regexp.MustCompile(`^([^:]+):(.+)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) != 3 {
		return nil, fmt.Errorf("\nInvalid box path format, should be BOX_ID:PATH")
	}
	return &BoxPath{
		BoxID: matches[1],
		Path:  matches[2],
	}, nil
}

// Check if path is a box path
func isBoxPath(path string) bool {
	return strings.Contains(path, ":")
}

// Convert relative path to absolute path
func getAbsolutePath(path string) string {
	if _, err := os.Stat(path); err == nil {
		absPath, err := filepath.Abs(path)
		if err == nil {
			return absPath
		}
	}

	dir := filepath.Dir(path)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return path
	}

	return filepath.Join(absDir, filepath.Base(path))
}

func NewBoxCpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "cp",
		Short:              "Copy files/folders between a box and the local filesystem",
		Long:               "Copy files/folders between a box and the local filesystem",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Help information
			if len(args) == 1 && (args[0] == "--help" || args[0] == "help") {
				printBoxCpHelp()
				return
			}

			// Parameter validation
			if len(args) != 2 {
				printBoxCpHelp()
				os.Exit(1)
			}

			src := args[0]
			dst := args[1]
			debugEnabled := os.Getenv("DEBUG") == "true"
			apiURL := "http://localhost:28080/api/v1"
			if envURL := os.Getenv("API_ENDPOINT"); envURL != "" {
				apiURL = envURL + "/api/v1"
			}

			// Debug log
			debug := func(msg string) {
				if debugEnabled {
					fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
				}
			}

			// Determine copy direction and process
			if isBoxPath(src) && !isBoxPath(dst) {
				// Copy from box to local
				boxPath, err := parseBoxPath(src)
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}

				debug(fmt.Sprintf("Box ID: %s", boxPath.BoxID))
				debug(fmt.Sprintf("Source path: %s", boxPath.Path))
				debug(fmt.Sprintf("Destination path: %s", dst))

				if dst == "-" {
					// Copy from box to stdout as tar stream
					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("Sending GET request to: %s", requestURL))

					resp, err := http.Get(requestURL)
					if err != nil {
						fmt.Println("Error: Failed to download from box")
						os.Exit(1)
					}
					defer resp.Body.Close()

					debug(fmt.Sprintf("HTTP response status code: %d", resp.StatusCode))

					if resp.StatusCode != http.StatusOK {
						fmt.Println("Error: Failed to download from box, HTTP status code:", resp.StatusCode)
						os.Exit(1)
					}

					_, err = io.Copy(os.Stdout, resp.Body)
					if err != nil {
						fmt.Println("Error: Failed to write to stdout")
						os.Exit(1)
					}
				} else {
					// Convert local path to absolute path
					dst = getAbsolutePath(dst)
					debug(fmt.Sprintf("Absolute destination path: %s", dst))

					// Copy from box to local file
					err := os.MkdirAll(filepath.Dir(dst), 0755)
					if err != nil {
						fmt.Printf("Error: Failed to create destination directory: %v\n", err)
						os.Exit(1)
					}

					// Download to temporary file
					tempFile, err := os.CreateTemp("", "gbox-cp-")
					if err != nil {
						fmt.Printf("Error: Failed to create temporary file: %v\n", err)
						os.Exit(1)
					}
					tempFilePath := tempFile.Name()
					defer os.Remove(tempFilePath)

					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("Sending GET request to: %s", requestURL))

					resp, err := http.Get(requestURL)
					if err != nil {
						fmt.Println("Error: Failed to download from box")
						os.Exit(1)
					}
					defer resp.Body.Close()

					debug(fmt.Sprintf("HTTP response status code: %d", resp.StatusCode))

					if resp.StatusCode != http.StatusOK {
						fmt.Println("Error: Failed to download from box, HTTP status code:", resp.StatusCode)
						os.Exit(1)
					}

					_, err = io.Copy(tempFile, resp.Body)
					tempFile.Close()
					if err != nil {
						fmt.Printf("Error: Failed to write to temporary file: %v\n", err)
						os.Exit(1)
					}

					// Check format and extract
					debug(fmt.Sprintf("Extracting archive to: %s", filepath.Dir(dst)))
					dstDir := filepath.Dir(dst)
					srcBaseName := filepath.Base(boxPath.Path)

					// Try to extract as gzip tar
					cmd := exec.Command("tar", "-xzf", tempFilePath, "-C", dstDir, srcBaseName)
					err = cmd.Run()
					if err != nil {
						// Try to extract as regular tar
						cmd = exec.Command("tar", "-xf", tempFilePath, "-C", dstDir, srcBaseName)
						err = cmd.Run()
						if err != nil {
							fmt.Printf("Error: Failed to extract archive: %v\n", err)
							os.Exit(1)
						}
					}

					fmt.Fprintf(os.Stderr, "Copied from box %s:%s to %s\n", boxPath.BoxID, boxPath.Path, dst)
				}
			} else if !isBoxPath(src) && isBoxPath(dst) {
				// Copy from local to box
				boxPath, err := parseBoxPath(dst)
				if err != nil {
					fmt.Println("Error: ", err)
					os.Exit(1)
				}

				debug(fmt.Sprintf("Box ID: %s", boxPath.BoxID))
				debug(fmt.Sprintf("Destination path: %s", boxPath.Path))
				debug(fmt.Sprintf("Source path: %s", src))

				if src == "-" {
					// Copy tar stream from stdin to box
					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("Sending PUT request to: %s", requestURL))

					req, err := http.NewRequest("PUT", requestURL, os.Stdin)
					if err != nil {
						fmt.Printf("Error: Failed to create request: %v\n", err)
						os.Exit(1)
					}

					req.Header.Set("Content-Type", "application/x-tar")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("Error: Failed to upload to box")
						os.Exit(1)
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						fmt.Println("Error: Failed to upload to box, HTTP status code:", resp.StatusCode)
						os.Exit(1)
					}

					fmt.Fprintf(os.Stderr, "Copied from stdin to box %s:%s\n", boxPath.BoxID, boxPath.Path)
				} else {
					// Convert local path to absolute path
					src = getAbsolutePath(src)
					debug(fmt.Sprintf("Absolute source path: %s", src))

					// Check if source file exists
					if _, err := os.Stat(src); os.IsNotExist(err) {
						fmt.Printf("Error: Source file or directory does not exist: %s\n", src)
						os.Exit(1)
					}

					// Copy from local file to box
					tempFile, err := os.CreateTemp("", "gbox-cp-")
					if err != nil {
						fmt.Printf("Error: Failed to create temporary file: %v\n", err)
						os.Exit(1)
					}
					tempFilePath := tempFile.Name()
					tempFile.Close()
					defer os.Remove(tempFilePath)

					// Create tar archive
					cmd := exec.Command("tar", "--no-xattrs", "-czf", tempFilePath, "-C", filepath.Dir(src), filepath.Base(src))
					err = cmd.Run()
					if err != nil {
						fmt.Printf("Error: Failed to create tar archive: %v\n", err)
						os.Exit(1)
					}
					debug(fmt.Sprintf("Created tar archive: %s", src))

					// Get file size
					fileInfo, err := os.Stat(tempFilePath)
					if err != nil {
						fmt.Printf("Error: Failed to get temporary file info: %v\n", err)
						os.Exit(1)
					}
					fileSize := fileInfo.Size()

					// Upload archive to box
					file, err := os.Open(tempFilePath)
					if err != nil {
						fmt.Printf("Error: Failed to open temporary file: %v\n", err)
						os.Exit(1)
					}
					defer file.Close()

					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("Sending PUT request to: %s", requestURL))

					req, err := http.NewRequest("PUT", requestURL, file)
					if err != nil {
						fmt.Printf("Error: Failed to create request: %v\n", err)
						os.Exit(1)
					}

					req.Header.Set("Content-Type", "application/x-tar")
					req.Header.Set("Content-Length", fmt.Sprintf("%d", fileSize))

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("Error: Failed to upload to box")
						os.Exit(1)
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						fmt.Println("Error: Failed to upload to box, HTTP status code:", resp.StatusCode)
						os.Exit(1)
					}

					fmt.Fprintf(os.Stderr, "Copied from %s to box %s:%s\n", src, boxPath.BoxID, boxPath.Path)
				}
			} else {
				fmt.Println("Error: Invalid path format. One path must be a box path (BOX_ID:PATH) and the other must be a local path")
				os.Exit(1)
			}
		},
	}

	return cmd
}

func printBoxCpHelp() {
	fmt.Println("Usage: gbox box cp <src> <dst>")
	fmt.Println()
	fmt.Println("Parameters:")
	fmt.Println("    <src>  Source path. Can be:")
	fmt.Println("           - Local file/directory path (e.g., ./local_file, /tmp/data)")
	fmt.Println("           - Box path in format BOX_ID:SRC_PATH (e.g., 550e8400-e29b-41d4-a716-446655440000:/work)")
	fmt.Println("           - \"-\" to read from stdin (must be a tar stream)")
	fmt.Println()
	fmt.Println("    <dst>  Destination path. Can be:")
	fmt.Println("           - Local file/directory path (e.g., /tmp/app_logs)")
	fmt.Println("           - Box path in format BOX_ID:DST_PATH (e.g., 550e8400-e29b-41d4-a716-446655440000:/work)")
	fmt.Println("           - \"-\" to write to stdout (as a tar stream)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("    # Copy local file to box")
	fmt.Println("    gbox box cp ./local_file 550e8400-e29b-41d4-a716-446655440000:/work")
	fmt.Println()
	fmt.Println("    # Copy from box to local")
	fmt.Println("    gbox box cp 550e8400-e29b-41d4-a716-446655440000:/var/logs/ /tmp/app_logs")
	fmt.Println()
	fmt.Println("    # Copy tar stream from stdin to box")
	fmt.Println("    tar czf - ./local_dir | gbox box cp - 550e8400-e29b-41d4-a716-446655440000:/work")
	fmt.Println()
	fmt.Println("    # Copy from box to stdout as tar stream")
	fmt.Println("    gbox box cp 550e8400-e29b-41d4-a716-446655440000:/etc/hosts - | tar xzf -")
	fmt.Println()
	fmt.Println("    # Copy directory from local to box")
	fmt.Println("    gbox box cp ./app_data 550e8400-e29b-41d4-a716-446655440000:/data/")
	fmt.Println()
	fmt.Println("    # Copy directory from box to local")
	fmt.Println("    gbox box cp 550e8400-e29b-41d4-a716-446655440000:/var/logs/ /tmp/app_logs/")
}

// This is a placeholder file
// TODO: Implement box_cp functionality

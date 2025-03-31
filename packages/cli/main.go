package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/babelcloud/gru-sandbox/packages/cli/cmd"
)

func main() {
	processArgs()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processArgs() {
	execName := filepath.Base(os.Args[0])

	if strings.HasPrefix(execName, "gbox-") {
		cmdParts := strings.SplitN(strings.TrimPrefix(execName, "gbox-"), "-", 2)

		if len(cmdParts) == 2 {
			os.Args = append([]string{os.Args[0], cmdParts[0], cmdParts[1]}, os.Args[1:]...)
		}
	}
}

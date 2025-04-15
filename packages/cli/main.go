package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/babelcloud/gbox/packages/cli/cmd"
)

var binaryPattern = regexp.MustCompile(`^gbox-(darwin|linux|windows)-(amd64|arm64|386)$`)

func main() {
	processArgs()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processArgs() {
	execName := filepath.Base(os.Args[0])

	if !strings.HasPrefix(execName, "gbox-") {
		return
	}

	if binaryPattern.MatchString(execName) {
		return
	}

	cmd := strings.TrimPrefix(execName, "gbox-")
	parts := strings.SplitN(cmd, "-", 2)

	newArgs := make([]string, 0, len(os.Args)+2)
	newArgs = append(newArgs, os.Args[0])
	newArgs = append(newArgs, parts...)
	newArgs = append(newArgs, os.Args[1:]...)

	os.Args = newArgs
}

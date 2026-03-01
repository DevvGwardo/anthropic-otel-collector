package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	outputDir := flag.String("output-dir", "", "Directory to write dashboard JSON files (stdout if empty)")
	flag.Parse()

	dash, err := buildDashboard()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build dashboard: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(dash, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal dashboard: %v\n", err)
		os.Exit(1)
	}

	if *outputDir == "" {
		fmt.Println(string(data))
		return
	}

	path := filepath.Join(*outputDir, "anthropic-claude-code-usage.json")
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write dashboard: %v\n", err)
		os.Exit(1)
	}
}

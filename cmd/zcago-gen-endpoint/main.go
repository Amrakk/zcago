package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amrakk/zcago/internal/strx"
)

func main() {
	name := flag.String("name", "", "Name of the endpoint to generate (e.g., MessagesSendImage)")
	flag.Parse()

	if *name == "" {
		fmt.Fprintln(os.Stderr, "Error: -name flag is required")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Generating endpoint code for: %s\n", *name)

	apiDir := "./api"
	if err := os.MkdirAll(apiDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating api directory: %v\n", err)
		os.Exit(1)
	}

	fileName := strx.PascalToSnake(*name) + ".go"
	filePath := filepath.Join(apiDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: file already exists: %s\n", filePath)
		os.Exit(1)
	}

	content := generateEndpointSkeleton(*name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created file: %s\n", filePath)
}

func generateEndpointSkeleton(name string) string {
	return fmt.Sprintf("// %s endpoint skeleton\n\npackage api", name)
}

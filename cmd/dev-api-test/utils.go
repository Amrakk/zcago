package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"log"
	"os"

	"github.com/Amrakk/zcago/model"
)

func metadataGetter(path string) (model.AttachmentMetadata, error) {
	// #nosec G304 — path is controlled by internal test context
	f, err := os.Open(path)
	if err != nil {
		return model.AttachmentMetadata{}, err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Printf("failed to close file %q: %v", path, cerr)
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return model.AttachmentMetadata{}, err
	}

	reader := bufio.NewReader(f)
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return model.AttachmentMetadata{}, err
	}

	return model.AttachmentMetadata{
		Size:   info.Size(),
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

func printJSON(title string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(title+":", string(b))
	return nil
}

func rootDir() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("get working dir failed:", err)
		return "."
	}
	return wd
}

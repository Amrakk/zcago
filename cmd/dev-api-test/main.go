package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"log"
	"os"
	"path/filepath"

	"github.com/Amrakk/zcago"
	API "github.com/Amrakk/zcago/api"
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

type App struct {
	zalo     zcago.Zalo
	credPath string
}

func main() {
	app := &App{
		zalo:     zcago.NewZalo(zcago.WithImageMetadataGetter(metadataGetter)),
		credPath: filepath.Join(rootDir(), "cmd", "credentials.json"),
	}

	ctx := context.Background()
	if err := app.run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) run(ctx context.Context) error {
	api, err := a.authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return a.runAPITest(ctx, api)
}

func (a *App) authenticate(ctx context.Context) (zcago.API, error) {
	cred := a.loadCredentials()
	if cred != nil && cred.IsValid() {
		return a.zalo.Login(ctx, *cred)
	}
	return a.zalo.LoginQR(ctx, nil, nil)
}

// ---- Edit only this function to try a different API call ----
func (a *App) runAPITest(ctx context.Context, api zcago.API) error {
	// Example: swap this line to any `api.<Method>(ctx, ...)`
	res, err := api.GetAccountInfo(ctx)
	if err != nil {
		return err
	}
	return printJSON("Result", res)
}

// ------------------------------------------------------------

func (a *App) loadCredentials() *zcago.Credentials {
	if _, err := os.Stat(a.credPath); os.IsNotExist(err) {
		return nil
	}
	raw, err := os.ReadFile(a.credPath)
	if err != nil {
		fmt.Printf("Warning: read credentials failed: %v\n", err)
		return nil
	}
	var c zcago.Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		fmt.Printf("Warning: parse credentials failed: %v\n", err)
		return nil
	}
	return &c
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

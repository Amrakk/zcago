package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amrakk/zcago"
)

type App struct {
	zalo     zcago.Zalo
	credPath string
}

func main() {
	app := &App{
		zalo:     zcago.NewZalo(),
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
	res, err := api.FetchAccountInfo(ctx)
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

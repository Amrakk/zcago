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
		zalo:     zcago.NewZalo(zcago.WithLogLevel(2)),
		credPath: filepath.Join(rootDir(), "examples", "credentials.json"),
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

	return a.displayAccountInfo(ctx, api)
}

func (a *App) authenticate(ctx context.Context) (zcago.API, error) {
	cred := a.loadCredentials()

	var api zcago.API
	var err error

	if cred != nil && cred.IsValid() {
		api, err = a.zalo.Login(ctx, *cred)
	} else {
		api, err = a.zalo.LoginQR(ctx, nil, nil)
	}

	return api, err
}

func (a *App) displayAccountInfo(ctx context.Context, api zcago.API) error {
	info, err := api.FetchAccountInfo(ctx)
	if err != nil {
		return fmt.Errorf("fetch account info failed: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal account info failed: %w", err)
	}

	fmt.Println("Account Info:", string(data))
	return nil
}

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

func rootDir() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("get working dir failed:", err)
		return "."
	}
	return wd
}

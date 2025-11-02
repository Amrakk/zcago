package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amrakk/zcago"
	API "github.com/Amrakk/zcago/api"
)

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

// ---- Edit only this function to try a different API call ----
func (a *App) runAPITest(ctx context.Context, api zcago.API) error {
	// Example: swap this line to any `api.<Method>(ctx, ...)`
	var _ API.AcceptFriendRequestFn
	res, err := api.GetAccountInfo(ctx)
	if err != nil {
		return err
	}

	return printJSON("Result", res)
}

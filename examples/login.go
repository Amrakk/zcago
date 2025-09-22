package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amrakk/zcago"
)

var credPath = filepath.Join(projectRoot(), "examples", "credentials.json")

func main() {
	ctx := context.Background()
	zalo := zcago.NewZalo(nil)
	cred := loadCredentials()

	var (
		api zcago.API
		err error
	)

	if isValidCredentials(cred) {
		api, err = zalo.Login(ctx, *cred)
	} else {
		api, err = zalo.LoginQR(ctx, nil, nil)
	}
	if err != nil {
		fmt.Println("login failed:", err)
		return
	}

	if cred == nil {
		if err := storeCredentials(ctx, api); err != nil {
			fmt.Println("save credentials failed:", err)
			return
		}
	}
}

func isValidCredentials(c *zcago.Credentials) bool {
	return c != nil && len(c.Imei) > 0 && c.Cookie.IsValid() && len(c.UserAgent) > 0
}

func loadCredentials() *zcago.Credentials {
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return nil
	}
	raw, err := os.ReadFile(credPath)
	if err != nil {
		fmt.Println("read credentials failed:", err)
		return nil
	}

	var c zcago.Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		fmt.Println("parse credentials failed:", err)
		return nil
	}
	return &c
}

func storeCredentials(ctx context.Context, api zcago.API) error {
	sc, err := api.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get context failed: %w", err)
	}

	lang := sc.Language()
	cookies := zcago.NewHTTPCookie(sc.Cookies())
	cred := zcago.NewCredentials(sc.IMEI(), cookies, sc.UserAgent(), &lang)

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials failed: %w", err)
	}

	if err := os.WriteFile(credPath, data, 0644); err != nil {
		return fmt.Errorf("write credentials failed: %w", err)
	}

	fmt.Println("Saved credentials to", credPath)
	return nil
}

func projectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("get working dir failed:", err)
		return "."
	}
	return wd
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Amrakk/zcago"
)

var credentialsPath = filepath.Join(getProjectRoot(), "examples", "credentials.json")

func main() {
	zalo := zcago.NewZalo(nil)
	ctx := context.Background()
	credentials := getCredentials()

	var api zcago.API
	var err error

	if credentials != nil && validateCredentials(credentials) {
		api, err = zalo.Login(ctx, *credentials)
	} else {
		api, err = zalo.LoginQR(ctx, nil, nil)
	}
	if err != nil {
		panic(err)
	}

	sc, err := api.GetContext(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Logged in as ", sc.LoginInfo.UID)

	if credentials == nil {
		saveCredentials(api, ctx)
	}
}

func validateCredentials(c *zcago.Credentials) bool {
	return c != nil && len(c.Imei) > 0 && c.Cookie.IsValid() && len(c.UserAgent) > 0
}

func getCredentials() *zcago.Credentials {
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return nil
	}

	raw, err := os.ReadFile(credentialsPath)
	if err != nil {
		panic(err)
	}

	var credentials zcago.Credentials
	if err := json.Unmarshal(raw, &credentials); err != nil {
		panic(err)
	}
	return &credentials
}

func saveCredentials(api zcago.API, ctx context.Context) {
	// sc, err := api.GetContext(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	// credentials := &zcago.Credentials{
	// 	Cookie:    sc.Cookie.ToJSON().Cookies,
	// 	Imei:      sc.Imei,
	// 	UserAgent: sc.UserAgent,
	// }

	// data, err := json.MarshalIndent(credentials, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }

	// if err := os.WriteFile(credentialsPath, data, 0644); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Saved credentials to", credentialsPath)
	panic("unimplemented")
}

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

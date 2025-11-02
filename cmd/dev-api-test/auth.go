package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Amrakk/zcago"
)

func (a *App) authenticate(ctx context.Context) (zcago.API, error) {
	cred := a.loadCredentials()
	if cred != nil && cred.IsValid() {
		return a.zalo.Login(ctx, *cred)
	}
	return a.zalo.LoginQR(ctx, nil, nil)
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

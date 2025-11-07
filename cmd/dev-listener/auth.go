package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Amrakk/zcago"
)

func (a *ListenerApp) authenticate() (zcago.API, error) {
	cred := a.loadCredentials()

	var api zcago.API
	var err error

	if cred != nil && cred.IsValid() {
		fmt.Println("Using saved credentials for login...")
		api, err = a.zalo.Login(a.ctx, *cred)
	} else {
		fmt.Println("No valid credentials found, starting QR login...")
		api, err = a.zalo.LoginQR(a.ctx, nil, nil)
	}

	return api, err
}

func (a *ListenerApp) loadCredentials() *zcago.Credentials {
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

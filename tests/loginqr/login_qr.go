package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Amrakk/zcago"
	"github.com/Amrakk/zcago/session/auth"
)

func main() {
	z := zcago.NewZalo()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiClient, err := z.LoginQR(ctx, &zcago.LoginQROption{}, func(ev auth.LoginQREvent) (any, error) {
		switch e := ev.(type) {
		case auth.EventQRCodeGenerated:
			fmt.Println("Scan this QR Code to login:")
			data := base64.StdEncoding.EncodeToString([]byte(e.Data.Image))

			fmt.Printf("\x1b_Ga=T,f=100;%s\x1b\\\n", data)

			if err := e.Actions.SaveToFile(ctx, ""); err != nil {
				log.Printf("save QR failed: %v", err)
			}

		case auth.EventQRCodeScanned:
			fmt.Println("QR code scanned, please confirm on your phone")

		case auth.EventQRCodeExpired:
			_ = e.Actions.Retry(ctx)

		case auth.EventQRCodeDeclined:
			_ = e.Actions.Abort(ctx)

		case auth.EventGotLoginInfo:
			fmt.Printf("Login info: IMEI=%s, UA=%s\n",
				e.Data.IMEI, e.Data.UserAgent)
		}
		return nil, nil
	})
	if err != nil {
		log.Fatalf("loginQR failed: %v", err)
	}

	info, err := apiClient.FetchAccountInfo(ctx)
	if err != nil {
		log.Fatalf("FetchAccountInfo failed: %v", err)
	}

	raw, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(raw))
}

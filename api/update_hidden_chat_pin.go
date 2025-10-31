package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"regexp"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

var ErrInvalidPIN = errs.NewZCA("invalid pin format", "api.UpdateHiddenChatPIN")

type (
	UpdateHiddenChatPINResponse = string
	UpdateHiddenChatPINFn       = func(ctx context.Context, pin string) (UpdateHiddenChatPINResponse, error)
)

func (a *api) UpdateHiddenChatPIN(ctx context.Context, pin string) (UpdateHiddenChatPINResponse, error) {
	return a.e.UpdateHiddenChatPIN(ctx, pin)
}

var updateHiddenChatPINFactory = apiFactory[UpdateHiddenChatPINResponse, UpdateHiddenChatPINFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateHiddenChatPINResponse]) (UpdateHiddenChatPINFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/hiddenconvers/update-pin", nil, true)

		regex := `^\d{4}$`

		return func(ctx context.Context, pin string) (UpdateHiddenChatPINResponse, error) {
			if !regexp.MustCompile(regex).MatchString(pin) {
				return "", ErrInvalidPIN
			}

			hash := md5.Sum([]byte(pin))
			encPIN := hex.EncodeToString(hash[:])

			payload := map[string]any{
				"new_pin": encPIN,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateHiddenChatPIN", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

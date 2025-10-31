package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	ResetHiddenChatPINResponse = string
	ResetHiddenChatPINFn       = func(ctx context.Context) (ResetHiddenChatPINResponse, error)
)

func (a *api) ResetHiddenChatPIN(ctx context.Context) (ResetHiddenChatPINResponse, error) {
	return a.e.ResetHiddenChatPIN(ctx)
}

var resetHiddenChatPINFactory = apiFactory[ResetHiddenChatPINResponse, ResetHiddenChatPINFn]()(
	func(a *api, sc session.Context, u factoryUtils[ResetHiddenChatPINResponse]) (ResetHiddenChatPINFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/hiddenconvers/reset", nil, true)

		return func(ctx context.Context) (ResetHiddenChatPINResponse, error) {
			payload := struct{}{}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.ResetHiddenChatPIN", err)
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

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
	UpdateActiveStatusResponse struct {
		Status bool `json:"status"`
	}
	UpdateActiveStatusFn = func(ctx context.Context, isActive bool) (*UpdateActiveStatusResponse, error)
)

func (a *api) UpdateActiveStatus(ctx context.Context, isActive bool) (*UpdateActiveStatusResponse, error) {
	return a.e.UpdateActiveStatus(ctx, isActive)
}

var updateActiveStatusFactory = apiFactory[*UpdateActiveStatusResponse, UpdateActiveStatusFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UpdateActiveStatusResponse]) (UpdateActiveStatusFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/profile", nil, false)

		return func(ctx context.Context, isActive bool) (*UpdateActiveStatusResponse, error) {
			payload := map[string]any{
				"status": jsonx.B2I(isActive),
				"imei":   sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateActiveStatus", err)
			}

			if isActive {
				serviceURL += "/ping"
			} else {
				serviceURL += "/deactive"
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

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
	LastOnlineSettings struct {
		ShowOnlineStatus bool `json:"show_online_status"`
	}
	LastOnlineResponse struct {
		Settings   LastOnlineSettings `json:"settings"`
		LastOnline int64              `json:"lastOnline"`
	}
	LastOnlineFn = func(ctx context.Context, userID string) (*LastOnlineResponse, error)
)

func (a *api) LastOnline(ctx context.Context, userID string) (*LastOnlineResponse, error) {
	return a.e.LastOnline(ctx, userID)
}

var lastOnlineFactory = apiFactory[*LastOnlineResponse, LastOnlineFn]()(
	func(a *api, sc session.Context, u factoryUtils[*LastOnlineResponse]) (LastOnlineFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/profile/lastOnline", nil, true)

		return func(ctx context.Context, userID string) (*LastOnlineResponse, error) {
			payload := map[string]any{
				"uid":       userID,
				"conv_type": 1,
				"imei":      sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.LastOnline", err)
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

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
	OnlineStatus struct {
		UserID string `json:"userId"`
		Status string `json:"status"`
	}

	GetFriendOnlineStatusResponse struct {
		Predefine   []string       `json:"predefine"`
		OwnerStatus string         `json:"ownerStatus"`
		Onlines     []OnlineStatus `json:"onlines"`
	}

	GetFriendOnlineStatusFn = func(ctx context.Context) (*GetFriendOnlineStatusResponse, error)
)

func (a *api) GetFriendOnlineStatus(ctx context.Context) (*GetFriendOnlineStatusResponse, error) {
	return a.e.GetFriendOnlineStatus(ctx)
}

var getFriendOnlineStatusFactory = apiFactory[*GetFriendOnlineStatusResponse, GetFriendOnlineStatusFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetFriendOnlineStatusResponse]) (GetFriendOnlineStatusFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/friend/onlines", nil, true)

		return func(ctx context.Context) (*GetFriendOnlineStatusResponse, error) {
			payload := map[string]any{
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetFriendOnlineStatus", err)
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

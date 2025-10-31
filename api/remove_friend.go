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
	RemoveFriendResponse = string
	RemoveFriendFn       = func(ctx context.Context, friendID string) (RemoveFriendResponse, error)
)

func (a *api) RemoveFriend(ctx context.Context, friendID string) (RemoveFriendResponse, error) {
	return a.e.RemoveFriend(ctx, friendID)
}

var removeFriendFactory = apiFactory[RemoveFriendResponse, RemoveFriendFn]()(
	func(a *api, sc session.Context, u factoryUtils[RemoveFriendResponse]) (RemoveFriendFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/remove", nil, true)

		return func(ctx context.Context, friendID string) (RemoveFriendResponse, error) {
			payload := map[string]any{
				"fid":  friendID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.RemoveFriend", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

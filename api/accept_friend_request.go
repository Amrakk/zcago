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
	AcceptFriendRequestResponse = string
	AcceptFriendRequestFn       = func(ctx context.Context, friendID string) (AcceptFriendRequestResponse, error)
)

func (a *api) AcceptFriendRequest(ctx context.Context, friendID string) (AcceptFriendRequestResponse, error) {
	return a.e.AcceptFriendRequest(ctx, friendID)
}

var acceptFriendRequestFactory = apiFactory[AcceptFriendRequestResponse, AcceptFriendRequestFn]()(
	func(a *api, sc session.Context, u factoryUtils[AcceptFriendRequestResponse]) (AcceptFriendRequestFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/accept", nil, true)

		return func(ctx context.Context, friendID string) (AcceptFriendRequestResponse, error) {
			payload := map[string]any{
				"fid":      friendID,
				"language": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.AcceptFriendRequest", err)
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

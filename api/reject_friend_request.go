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
	RejectFriendRequestResponse = string
	RejectFriendRequestFn       = func(ctx context.Context, friendID string) (RejectFriendRequestResponse, error)
)

func (a *api) RejectFriendRequest(ctx context.Context, friendID string) (RejectFriendRequestResponse, error) {
	return a.e.RejectFriendRequest(ctx, friendID)
}

var rejectFriendRequestFactory = apiFactory[RejectFriendRequestResponse, RejectFriendRequestFn]()(
	func(a *api, sc session.Context, u factoryUtils[RejectFriendRequestResponse]) (RejectFriendRequestFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/reject", nil, true)

		return func(ctx context.Context, friendID string) (RejectFriendRequestResponse, error) {
			payload := map[string]any{
				"fid": friendID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.RejectFriendRequests", err)
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

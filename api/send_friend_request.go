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
	SendFriendRequestResponse = string
	SendFriendRequestFn       = func(ctx context.Context, message string, userID string) (SendFriendRequestResponse, error)
)

func (a *api) SendFriendRequest(ctx context.Context, message string, userID string) (SendFriendRequestResponse, error) {
	return a.e.SendFriendRequest(ctx, message, userID)
}

var sendFriendRequestFactory = apiFactory[SendFriendRequestResponse, SendFriendRequestFn]()(
	func(a *api, sc session.Context, u factoryUtils[SendFriendRequestResponse]) (SendFriendRequestFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/sendreq", nil, true)

		return func(ctx context.Context, message string, userID string) (SendFriendRequestResponse, error) {
			payload := map[string]any{
				"toid":     userID,
				"msg":      message,
				"reqsrc":   30,
				"imei":     sc.IMEI(),
				"language": sc.Language(),
				"srcParams": jsonx.Stringify(map[string]any{
					"uidTo": userID,
				}),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SendFriendRequest", err)
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

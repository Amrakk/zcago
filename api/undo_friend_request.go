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
	UndoFriendRequestResponse = string
	UndoFriendRequestFn       = func(ctx context.Context, userID string) (UndoFriendRequestResponse, error)
)

func (a *api) UndoFriendRequest(ctx context.Context, userID string) (UndoFriendRequestResponse, error) {
	return a.e.UndoFriendRequest(ctx, userID)
}

var undoFriendRequestFactory = apiFactory[UndoFriendRequestResponse, UndoFriendRequestFn]()(
	func(a *api, sc session.Context, u factoryUtils[UndoFriendRequestResponse]) (UndoFriendRequestFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/undo", nil, true)

		return func(ctx context.Context, userID string) (UndoFriendRequestResponse, error) {
			payload := map[string]any{
				"fid": userID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UndoFriendRequest", err)
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

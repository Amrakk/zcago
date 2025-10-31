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
	BlockUserResponse = string
	BlockUserFn       = func(ctx context.Context, userID string) (BlockUserResponse, error)
)

func (a *api) BlockUser(ctx context.Context, userID string) (BlockUserResponse, error) {
	return a.e.BlockUser(ctx, userID)
}

var blockUserFactory = apiFactory[BlockUserResponse, BlockUserFn]()(
	func(a *api, sc session.Context, u factoryUtils[BlockUserResponse]) (BlockUserFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/block", nil, true)

		return func(ctx context.Context, userID string) (BlockUserResponse, error) {
			payload := map[string]any{
				"fid":  userID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.BlockUser", err)
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

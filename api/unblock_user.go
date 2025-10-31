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
	UnblockUserResponse = string
	UnblockUserFn       = func(ctx context.Context, userID string) (UnblockUserResponse, error)
)

func (a *api) UnblockUser(ctx context.Context, userID string) (UnblockUserResponse, error) {
	return a.e.UnblockUser(ctx, userID)
}

var unblockUserFactory = apiFactory[UnblockUserResponse, UnblockUserFn]()(
	func(a *api, sc session.Context, u factoryUtils[UnblockUserResponse]) (UnblockUserFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/unblock", nil, true)

		return func(ctx context.Context, userID string) (UnblockUserResponse, error) {
			payload := map[string]any{
				"fid":  userID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UnblockUser", err)
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

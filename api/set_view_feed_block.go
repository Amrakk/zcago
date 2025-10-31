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
	SetViewFeedBlockResponse = string
	SetViewFeedBlockFn       = func(ctx context.Context, userID string, isBlock bool) (SetViewFeedBlockResponse, error)
)

func (a *api) SetViewFeedBlock(ctx context.Context, userID string, isBlock bool) (SetViewFeedBlockResponse, error) {
	return a.e.SetViewFeedBlock(ctx, userID, isBlock)
}

var setViewFeedBlockFactory = apiFactory[SetViewFeedBlockResponse, SetViewFeedBlockFn]()(
	func(a *api, sc session.Context, u factoryUtils[SetViewFeedBlockResponse]) (SetViewFeedBlockFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/feed/block", nil, true)

		return func(ctx context.Context, userID string, isBlock bool) (SetViewFeedBlockResponse, error) {
			payload := map[string]any{
				"fid":         userID,
				"isBlockFeed": jsonx.B2I(isBlock),
				"imei":        sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SetViewFeedBlock", err)
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

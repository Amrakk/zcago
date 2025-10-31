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
	RemoveAliasResponse = string
	RemoveAliasFn       = func(ctx context.Context, userID string) (RemoveAliasResponse, error)
)

func (a *api) RemoveAlias(ctx context.Context, userID string) (RemoveAliasResponse, error) {
	return a.e.RemoveAlias(ctx, userID)
}

var removeAliasFactory = apiFactory[RemoveAliasResponse, RemoveAliasFn]()(
	func(a *api, sc session.Context, u factoryUtils[RemoveAliasResponse]) (RemoveAliasFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("alias"), "")
		serviceURL := u.MakeURL(base+"/api/alias/remove", nil, true)

		return func(ctx context.Context, userID string) (RemoveAliasResponse, error) {
			payload := map[string]any{
				"friendId": userID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.RemoveAlias", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

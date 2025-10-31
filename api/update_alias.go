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
	UpdateAliasResponse = string
	UpdateAliasFn       = func(ctx context.Context, userID string, alias string) (UpdateAliasResponse, error)
)

func (a *api) UpdateAlias(ctx context.Context, userID string, alias string) (UpdateAliasResponse, error) {
	return a.e.UpdateAlias(ctx, userID, alias)
}

var updateAliasFactory = apiFactory[UpdateAliasResponse, UpdateAliasFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateAliasResponse]) (UpdateAliasFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("alias"), "")
		serviceURL := u.MakeURL(base+"/api/alias/update", nil, true)

		return func(ctx context.Context, userID string, alias string) (UpdateAliasResponse, error) {
			payload := map[string]any{
				"friendId": userID,
				"alias":    alias,
				"imei":     sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateAlias", err)
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

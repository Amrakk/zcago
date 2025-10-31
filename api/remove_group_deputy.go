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
	RemoveGroupDeputyResponse = string
	RemoveGroupDeputyFn       = func(ctx context.Context, groupID string, memberID ...string) (RemoveGroupDeputyResponse, error)
)

func (a *api) RemoveGroupDeputy(ctx context.Context, groupID string, memberID ...string) (RemoveGroupDeputyResponse, error) {
	return a.e.RemoveGroupDeputy(ctx, groupID, memberID...)
}

var removeGroupDeputyFactory = apiFactory[RemoveGroupDeputyResponse, RemoveGroupDeputyFn]()(
	func(a *api, sc session.Context, u factoryUtils[RemoveGroupDeputyResponse]) (RemoveGroupDeputyFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/admins/remove", nil, true)

		return func(ctx context.Context, groupID string, memberID ...string) (RemoveGroupDeputyResponse, error) {
			payload := map[string]any{
				"grid":    groupID,
				"members": memberID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.RemoveGroupDeputy", err)
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

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
	AddGroupDeputyResponse = string
	AddGroupDeputyFn       = func(ctx context.Context, groupID string, memberID ...string) (AddGroupDeputyResponse, error)
)

func (a *api) AddGroupDeputy(ctx context.Context, groupID string, memberID ...string) (AddGroupDeputyResponse, error) {
	return a.e.AddGroupDeputy(ctx, groupID, memberID...)
}

var addGroupDeputyFactory = apiFactory[AddGroupDeputyResponse, AddGroupDeputyFn]()(
	func(a *api, sc session.Context, u factoryUtils[AddGroupDeputyResponse]) (AddGroupDeputyFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/admins/add", nil, true)

		return func(ctx context.Context, groupID string, memberID ...string) (AddGroupDeputyResponse, error) {
			payload := map[string]any{
				"grid":    groupID,
				"members": memberID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.AddGroupDeputy", err)
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

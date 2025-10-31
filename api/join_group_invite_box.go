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
	JoinGroupInviteBoxResponse = string
	JoinGroupInviteBoxFn       = func(ctx context.Context, groupID string) (JoinGroupInviteBoxResponse, error)
)

func (a *api) JoinGroupInviteBox(ctx context.Context, groupID string) (JoinGroupInviteBoxResponse, error) {
	return a.e.JoinGroupInviteBox(ctx, groupID)
}

var joinGroupInviteBoxFactory = apiFactory[JoinGroupInviteBoxResponse, JoinGroupInviteBoxFn]()(
	func(a *api, sc session.Context, u factoryUtils[JoinGroupInviteBoxResponse]) (JoinGroupInviteBoxFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/inv-box/join", nil, true)

		return func(ctx context.Context, groupID string) (JoinGroupInviteBoxResponse, error) {
			payload := map[string]any{
				"grid": groupID,
				"lang": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.JoinGroupInviteBox", err)
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

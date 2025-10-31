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
	DisableGroupLinkResponse = string
	DisableGroupLinkFn       = func(ctx context.Context, groupID string) (DisableGroupLinkResponse, error)
)

func (a *api) DisableGroupLink(ctx context.Context, groupID string) (DisableGroupLinkResponse, error) {
	return a.e.DisableGroupLink(ctx, groupID)
}

var disableGroupLinkFactory = apiFactory[DisableGroupLinkResponse, DisableGroupLinkFn]()(
	func(a *api, sc session.Context, u factoryUtils[DisableGroupLinkResponse]) (DisableGroupLinkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/link/disable", nil, true)

		return func(ctx context.Context, groupID string) (DisableGroupLinkResponse, error) {
			payload := map[string]any{
				"grid": groupID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.DisableGroupLink", err)
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

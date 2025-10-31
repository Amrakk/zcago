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
	EnableGroupLinkResponse struct {
		Link           string `json:"link"`
		ExpirationDate int64  `json:"expiration_date"`
		Enabled        int    `json:"enabled"`
	}
	EnableGroupLinkFn = func(ctx context.Context, groupID string) (*EnableGroupLinkResponse, error)
)

func (a *api) EnableGroupLink(ctx context.Context, groupID string) (*EnableGroupLinkResponse, error) {
	return a.e.EnableGroupLink(ctx, groupID)
}

var enableGroupLinkFactory = apiFactory[*EnableGroupLinkResponse, EnableGroupLinkFn]()(
	func(a *api, sc session.Context, u factoryUtils[*EnableGroupLinkResponse]) (EnableGroupLinkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/link/new", nil, true)

		return func(ctx context.Context, groupID string) (*EnableGroupLinkResponse, error) {
			payload := map[string]any{
				"grid": groupID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.EnableGroupLink", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

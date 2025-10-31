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
	GetGroupLinkDetailResponse struct {
		Link           *string `json:"link"`
		ExpirationDate *int64  `json:"expirationDate"`
		Enabled        int     `json:"enabled"` // 1: enabled, 0: disabled
	}
	GetGroupLinkDetailFn = func(ctx context.Context, groupID string) (*GetGroupLinkDetailResponse, error)
)

func (a *api) GetGroupLinkDetail(ctx context.Context, groupID string) (*GetGroupLinkDetailResponse, error) {
	return a.e.GetGroupLinkDetail(ctx, groupID)
}

var getGroupLinkDetailFactory = apiFactory[*GetGroupLinkDetailResponse, GetGroupLinkDetailFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupLinkDetailResponse]) (GetGroupLinkDetailFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/link/detail", nil, true)

		return func(ctx context.Context, groupID string) (*GetGroupLinkDetailResponse, error) {
			payload := map[string]any{
				"grid": groupID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupLinkDetail", err)
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

package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	GetAliasListItem struct {
		UserID string `json:"userid"`
		Alias  string `json:"alias"`
	}
	GetAliasListResponse struct {
		Items      []GetAliasListItem `json:"items"`
		UpdateTime string             `json:"updateTime"`
	}
	GetAliasListFn = func(ctx context.Context, options model.OffsetPaginationOptions) (*GetAliasListResponse, error)
)

func (a *api) GetAliasList(ctx context.Context, options model.OffsetPaginationOptions) (*GetAliasListResponse, error) {
	return a.e.GetAliasList(ctx, options)
}

var getAliasListFactory = apiFactory[*GetAliasListResponse, GetAliasListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAliasListResponse]) (GetAliasListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("alias"), "")
		serviceURL := u.MakeURL(base+"/api/alias/list", nil, true)

		return func(ctx context.Context, options model.OffsetPaginationOptions) (*GetAliasListResponse, error) {
			if options.Count <= 0 {
				options.Count = 100
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			payload := map[string]any{
				"page":  options.Page,
				"count": options.Count,
				"imei":  sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetAliasList", err)
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

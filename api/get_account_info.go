package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	GetAccountInfoResponse struct {
		Profile model.User `json:"profile"`
	}
	GetAccountInfoFn = func(ctx context.Context) (*GetAccountInfoResponse, error)
)

func (a *api) GetAccountInfo(ctx context.Context) (*GetAccountInfoResponse, error) {
	return a.e.GetAccountInfo(ctx)
}

var getAccountInfoFactory = apiFactory[*GetAccountInfoResponse, GetAccountInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAccountInfoResponse]) (GetAccountInfoFn, error) {
		baseURL := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		url := u.MakeURL(baseURL+"/api/social/profile/me-v2", nil, true)

		return func(ctx context.Context) (*GetAccountInfoResponse, error) {
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

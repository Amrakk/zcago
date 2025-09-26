package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type FetchAccountInfoResponse struct {
	Profile model.User `json:"profile"`
}
type FetchAccountInfoFn = func(ctx context.Context) (*FetchAccountInfoResponse, error)

func (a *api) FetchAccountInfo(ctx context.Context) (*FetchAccountInfoResponse, error) {
	return a.e.FetchAccountInfo(ctx)
}

var fetchAccountInfoFactory = apiFactory[*FetchAccountInfoResponse, FetchAccountInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*FetchAccountInfoResponse]) (FetchAccountInfoFn, error) {
		baseURL := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		url := u.MakeURL(baseURL+"/api/social/profile/me-v2", nil, true)

		return func(ctx context.Context) (*FetchAccountInfoResponse, error) {
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, nil, true)
		}, nil
	},
)

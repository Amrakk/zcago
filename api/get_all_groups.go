package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	GroupIDVerMap        map[string]string
	GetAllGroupsResponse struct {
		Version    string        `json:"version"`
		GridVerMap GroupIDVerMap `json:"gridVerMap"`
	}
	GetAllGroupsFn = func(ctx context.Context) (*GetAllGroupsResponse, error)
)

func (a *api) GetAllGroups(ctx context.Context) (*GetAllGroupsResponse, error) {
	return a.e.GetAllGroups(ctx)
}

var getAllGroupsFactory = apiFactory[*GetAllGroupsResponse, GetAllGroupsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAllGroupsResponse]) (GetAllGroupsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_poll"), "")
		serviceURL := u.MakeURL(base+"/api/group/getlg/v4", nil, true)

		return func(ctx context.Context) (*GetAllGroupsResponse, error) {
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

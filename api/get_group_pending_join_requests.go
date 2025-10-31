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
	RequestUserInfo struct {
		UID    string `json:"uid"`
		DPN    string `json:"dpn"`
		Avatar string `json:"avatar"`
	}

	GetGroupPendingJoinRequestsResponse struct {
		Users []RequestUserInfo `json:"users"`
		Time  int64             `json:"time"`
	}
	GetGroupPendingJoinRequestsFn = func(ctx context.Context, groupID string) (*GetGroupPendingJoinRequestsResponse, error)
)

func (a *api) GetGroupPendingJoinRequests(ctx context.Context, groupID string) (*GetGroupPendingJoinRequestsResponse, error) {
	return a.e.GetGroupPendingJoinRequests(ctx, groupID)
}

var getGroupPendingJoinRequestsFactory = apiFactory[*GetGroupPendingJoinRequestsResponse, GetGroupPendingJoinRequestsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupPendingJoinRequestsResponse]) (GetGroupPendingJoinRequestsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/pending-mems/list", nil, true)

		return func(ctx context.Context, groupID string) (*GetGroupPendingJoinRequestsResponse, error) {
			payload := map[string]any{
				"grid": groupID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupPendingJoinRequests", err)
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

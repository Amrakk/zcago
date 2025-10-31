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
	GetFriendRequestStatusResponse struct {
		AddFriendPrivacy int  `json:"addFriendPrivacy"`
		IsSeenFriendReq  bool `json:"isSeenFriendReq"`
		IsFriend         int  `json:"is_friend"`
		IsRequested      int  `json:"is_requested"`
		IsRequesting     int  `json:"is_requesting"`
	}
	GetFriendRequestStatusFn = func(ctx context.Context, friendID string) (*GetFriendRequestStatusResponse, error)
)

func (a *api) GetFriendRequestStatus(ctx context.Context, friendID string) (*GetFriendRequestStatusResponse, error) {
	return a.e.GetFriendRequestStatus(ctx, friendID)
}

var getFriendRequestStatusFactory = apiFactory[*GetFriendRequestStatusResponse, GetFriendRequestStatusFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetFriendRequestStatusResponse]) (GetFriendRequestStatusFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/reqstatus", nil, true)

		return func(ctx context.Context, friendID string) (*GetFriendRequestStatusResponse, error) {
			payload := map[string]any{
				"fid":  friendID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetFriendRequestStatus", err)
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

package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	UserProfile struct {
		DisplayName    string `json:"displayName"`
		ZaloName       string `json:"zaloName"`
		Avatar         string `json:"avatar"`
		AccountStatus  int    `json:"accountStatus"`
		Type           int    `json:"type"`
		LastUpdateTime int64  `json:"lastUpdateTime"`
		ID             string `json:"id"`
	}

	GetUserSummaryInfoResponse struct {
		Profiles          map[string]UserProfile `json:"profiles"`
		UnchangedsProfile []any                  `json:"unchangeds_profile"`
	}
	GetUserSummaryInfoFn = func(ctx context.Context, userID ...string) (*GetUserSummaryInfoResponse, error)
)

func (a *api) GetUserSummaryInfo(ctx context.Context, userID ...string) (*GetUserSummaryInfoResponse, error) {
	return a.e.GetUserSummaryInfo(ctx, userID...)
}

var getUserSummaryInfoFactory = apiFactory[*GetUserSummaryInfoResponse, GetUserSummaryInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetUserSummaryInfoResponse]) (GetUserSummaryInfoFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/group/members", nil, true)

		return func(ctx context.Context, userID ...string) (*GetUserSummaryInfoResponse, error) {
			idMap := make([]string, len(userID))
			for i, id := range userID {
				if !strings.HasSuffix(id, "_0") {
					id += "_0"
				}
				idMap[i] = id
			}

			payload := map[string]any{
				"friend_pversion_map": idMap,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetUserSummaryInfo", err)
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

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
	GroupInfoPendingApprove struct {
		Time int64    `json:"time"`
		UIDs []string `json:"uids"`
	}
	GroupInfo struct {
		model.GroupInfo
		MemVerList     []string                `json:"memVerList"`
		PendingApprove GroupInfoPendingApprove `json:"pendingApprove"`
	}

	GetGroupInfoResponse struct {
		RemovedsGroup   []string             `json:"removedsGroup"`
		UnchangedsGroup []string             `json:"unchangedsGroup"`
		GridInfoMap     map[string]GroupInfo `json:"gridInfoMap"`
	}
	GetGroupInfoFn = func(ctx context.Context, groupID ...string) (*GetGroupInfoResponse, error)
)

func (a *api) GetGroupInfo(ctx context.Context, groupID ...string) (*GetGroupInfoResponse, error) {
	return a.e.GetGroupInfo(ctx, groupID...)
}

var getGroupInfoFactory = apiFactory[*GetGroupInfoResponse, GetGroupInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupInfoResponse]) (GetGroupInfoFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/getmg-v2", nil, true)

		return func(ctx context.Context, groupID ...string) (*GetGroupInfoResponse, error) {
			gridVerMap := make(map[string]int, len(groupID))
			for _, id := range groupID {
				gridVerMap[id] = 0
			}

			payload := map[string]any{
				"gridVerMap": jsonx.Stringify(gridVerMap),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupInfo", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	GroupInviteBoxInfoTopic struct {
		model.GroupTopic
		Action *int `json:"action,omitempty"`
	}
	GroupInviteBoxInfo struct {
		model.GroupInfo
		Topic *GroupInviteBoxInfoTopic `json:"topic,omitempty"`
	}

	GetGroupInviteBoxInfoResponse struct {
		GroupInfo        GroupInviteBoxInfo `json:"groupInfo"`
		InviterInfo      model.UserSummary  `json:"inviterInfo"`
		GroupCreatorInfo model.UserSummary  `json:"grCreatorInfo"`
		ExpiredTS        string             `json:"expiredTs"`
		Type             int                `json:"type"`
	}
	GetGroupInviteBoxInfoFn = func(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupInviteBoxInfoResponse, error)
)

func (a *api) GetGroupInviteBoxInfo(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupInviteBoxInfoResponse, error) {
	return a.e.GetGroupInviteBoxInfo(ctx, groupID, options)
}

var getGroupInviteBoxInfoFactory = apiFactory[*GetGroupInviteBoxInfoResponse, GetGroupInviteBoxInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupInviteBoxInfoResponse]) (GetGroupInviteBoxInfoFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/inv-box/inv-info", nil, true)

		return func(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupInviteBoxInfoResponse, error) {
			if options.Count <= 0 {
				options.Count = 10
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			payload := map[string]any{
				"grId":   groupID,
				"mcount": options.Count,
				"mpage":  options.Page,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupInviteBoxInfo", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

func (r *GetGroupInviteBoxInfoResponse) UnmarshalJSON(data []byte) error {
	type alias GetGroupInviteBoxInfoResponse
	aux := &struct {
		*alias
	}{
		alias: (*alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

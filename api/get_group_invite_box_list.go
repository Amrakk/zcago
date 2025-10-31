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
	Invitation struct {
		GroupInfo        model.GroupInfo   `json:"groupInfo"`
		InviterInfo      model.UserSummary `json:"inviterInfo"`
		GroupCreatorInfo model.UserSummary `json:"grCreatorInfo"`
		ExpiredTS        string            `json:"expiredTs"` // Expired timestamp max 7 days
		Type             int               `json:"type"`
	}

	GetGroupInviteBoxListOptions struct {
		MPage      int
		Page       int
		InvPerPage int
		MCount     int
		// LastGroupID string
	}
	GetGroupInviteBoxListResponse struct {
		Invitations []Invitation `json:"invitations"`
		Total       int          `json:"total"`
		HasMore     bool         `json:"hasMore"`
	}
	GetGroupInviteBoxListFn = func(ctx context.Context, options GetGroupInviteBoxListOptions) (*GetGroupInviteBoxListResponse, error)
)

func (a *api) GetGroupInviteBoxList(ctx context.Context, options GetGroupInviteBoxListOptions) (*GetGroupInviteBoxListResponse, error) {
	return a.e.GetGroupInviteBoxList(ctx, options)
}

var getGroupInviteBoxListFactory = apiFactory[*GetGroupInviteBoxListResponse, GetGroupInviteBoxListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupInviteBoxListResponse]) (GetGroupInviteBoxListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/inv-box/list", nil, true)

		return func(ctx context.Context, options GetGroupInviteBoxListOptions) (*GetGroupInviteBoxListResponse, error) {
			if options.MCount <= 0 {
				options.MCount = 10
			}
			if options.MPage <= 0 {
				options.MPage = 1
			}
			if options.InvPerPage <= 0 {
				options.InvPerPage = 12
			}

			payload := map[string]any{
				"page":       options.Page,
				"invPerPage": options.InvPerPage,
				"mcount":     options.MCount,
				"mpage":      options.MPage,
				// "lastGroupId":        options.LastGroupID, // @TODO: check this behavior
				"lastGroupId":        nil,
				"avatar_size":        120,
				"member_avatar_size": 120,
			}

			// if len(options.LastGroupID) == 0 {
			// 	payload["lastGroupId"] = nil
			// }

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupInviteBoxList", err)
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

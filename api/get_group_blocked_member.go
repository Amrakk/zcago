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
	GetGroupBlockedMemberResponse struct {
		BlockedMembers []model.UserSummary `json:"blocked_members"`
		HasMore        int                 `json:"has_more"`
	}
	GetGroupBlockedMemberFn = func(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupBlockedMemberResponse, error)
)

func (a *api) GetGroupBlockedMember(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupBlockedMemberResponse, error) {
	return a.e.GetGroupBlockedMember(ctx, groupID, options)
}

var getGroupBlockedMemberFactory = apiFactory[*GetGroupBlockedMemberResponse, GetGroupBlockedMemberFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupBlockedMemberResponse]) (GetGroupBlockedMemberFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/blockedmems/list", nil, true)

		return func(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupBlockedMemberResponse, error) {
			if options.Count <= 0 {
				options.Count = 50
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			payload := map[string]any{
				"grid":  groupID,
				"page":  options.Page,
				"count": options.Count,
				"imei":  sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupBlockedMember", err)
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

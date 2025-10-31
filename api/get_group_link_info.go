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
	GetGroupLinkInfoResponse struct {
		GroupID       string              `json:"groupId"`
		Name          string              `json:"name"`
		Desc          string              `json:"desc"`
		Type          int                 `json:"type"`
		CreatorID     string              `json:"creatorId"`
		Avt           string              `json:"avt"`
		FullAvt       string              `json:"fullAvt"`
		AdminIDs      []string            `json:"adminIds"`
		CurrentMems   []model.UserSummary `json:"currentMems"`
		Admins        []any               `json:"admins"`
		HasMoreMember int                 `json:"hasMoreMember"`
		SubType       int                 `json:"subType"`
		TotalMember   int                 `json:"totalMember"`
		Setting       model.GroupSetting  `json:"setting"`
		GlobalID      string              `json:"globalId"`
	}
	GetGroupLinkInfoFn = func(ctx context.Context, link string, memberPage int) (*GetGroupLinkInfoResponse, error)
)

func (a *api) GetGroupLinkInfo(ctx context.Context, link string, memberPage int) (*GetGroupLinkInfoResponse, error) {
	return a.e.GetGroupLinkInfo(ctx, link, memberPage)
}

var getGroupLinkInfoFactory = apiFactory[*GetGroupLinkInfoResponse, GetGroupLinkInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupLinkInfoResponse]) (GetGroupLinkInfoFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/link/ginfo", nil, true)

		return func(ctx context.Context, link string, memberPage int) (*GetGroupLinkInfoResponse, error) {
			if memberPage <= 0 {
				memberPage = 1
			}

			payload := map[string]any{
				"link":               link,
				"avatar_size":        120,
				"member_avatar_size": 120,
				"mpage":              memberPage,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupLinkInfo", err)
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

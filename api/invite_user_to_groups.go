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
	InviteUserToGroupsMessage struct {
		ErrorCode    int     `json:"error_code"`
		ErrorMessage string  `json:"error_message"`
		Data         *string `json:"data"`
	}
	InviteUserToGroupsResponse struct {
		GridMessageMap map[string]InviteUserToGroupsMessage `json:"grid_message_map"`
	}
	InviteUserToGroupsFn = func(ctx context.Context, userID string, groupID ...string) (*InviteUserToGroupsResponse, error)
)

func (a *api) InviteUserToGroups(ctx context.Context, userID string, groupID ...string) (*InviteUserToGroupsResponse, error) {
	return a.e.InviteUserToGroups(ctx, userID, groupID...)
}

var inviteUserToGroupsFactory = apiFactory[*InviteUserToGroupsResponse, InviteUserToGroupsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*InviteUserToGroupsResponse]) (InviteUserToGroupsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/invite/multi", nil, true)

		return func(ctx context.Context, userID string, groupID ...string) (*InviteUserToGroupsResponse, error) {
			payload := map[string]any{
				"grids":          groupID,
				"member":         userID,
				"memberType":     -1,
				"srcInteraction": 2,
				"clientLang":     sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateLanguage", err)
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

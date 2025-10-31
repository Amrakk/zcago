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
	RemoveUserFromGroupResponse struct {
		ErrorMembers []string `json:"errorMembers"`
	}
	RemoveUserFromGroupFn = func(ctx context.Context, groupID string, memberID ...string) (*RemoveUserFromGroupResponse, error)
)

func (a *api) RemoveUserFromGroup(ctx context.Context, groupID string, memberID ...string) (*RemoveUserFromGroupResponse, error) {
	return a.e.RemoveUserFromGroup(ctx, groupID, memberID...)
}

var removeUserFromGroupFactory = apiFactory[*RemoveUserFromGroupResponse, RemoveUserFromGroupFn]()(
	func(a *api, sc session.Context, u factoryUtils[*RemoveUserFromGroupResponse]) (RemoveUserFromGroupFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/kickout", nil, true)

		return func(ctx context.Context, groupID string, memberID ...string) (*RemoveUserFromGroupResponse, error) {
			payload := map[string]any{
				"grid":    groupID,
				"members": memberID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.RemoveUserFromGroup", err)
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

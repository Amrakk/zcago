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
	AddUserToGroupResponse struct {
		ErrorMembers []string            `json:"errorMembers"`
		ErrorData    map[string][]string `json:"error_data"`
	}
	AddUserToGroupFn = func(ctx context.Context, groupID string, userID ...string) (*AddUserToGroupResponse, error)
)

func (a *api) AddUserToGroup(ctx context.Context, groupID string, userID ...string) (*AddUserToGroupResponse, error) {
	return a.e.AddUserToGroup(ctx, groupID, userID...)
}

var addUserToGroupFactory = apiFactory[*AddUserToGroupResponse, AddUserToGroupFn]()(
	func(a *api, sc session.Context, u factoryUtils[*AddUserToGroupResponse]) (AddUserToGroupFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/invite/v2", nil, true)

		return func(ctx context.Context, groupID string, userID ...string) (*AddUserToGroupResponse, error) {
			memberTypes := make([]int, len(userID))
			for i := range userID {
				memberTypes[i] = -1
			}

			payload := map[string]any{
				"grid":        groupID,
				"members":     userID,
				"memberTypes": memberTypes,
				"imei":        sc.IMEI(),
				"clientLang":  sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.AddUserToGroup", err)
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

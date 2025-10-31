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
	LeaveGroupResponse struct {
		MemberError []any `json:"memberError"`
	}
	LeaveGroupFn = func(ctx context.Context, groupID string, isSilent bool) (*LeaveGroupResponse, error)
)

func (a *api) LeaveGroup(ctx context.Context, groupID string, isSilent bool) (*LeaveGroupResponse, error) {
	return a.e.LeaveGroup(ctx, groupID, isSilent)
}

var leaveGroupFactory = apiFactory[*LeaveGroupResponse, LeaveGroupFn]()(
	func(a *api, sc session.Context, u factoryUtils[*LeaveGroupResponse]) (LeaveGroupFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/leave", nil, true)

		return func(ctx context.Context, groupID string, isSilent bool) (*LeaveGroupResponse, error) {
			payload := map[string]any{
				"grids":    []any{groupID}, // API only supports leaving one group at a time
				"silent":   jsonx.B2I(isSilent),
				"imei":     sc.IMEI(),
				"language": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.LeaveGroup", err)
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

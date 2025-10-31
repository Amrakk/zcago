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
	UpdateGroupNameResponse struct {
		Status int `json:"status"`
	}
	UpdateGroupNameFn = func(ctx context.Context, groupID string, name string) (*UpdateGroupNameResponse, error)
)

func (a *api) UpdateGroupName(ctx context.Context, groupID string, name string) (*UpdateGroupNameResponse, error) {
	return a.e.UpdateGroupName(ctx, groupID, name)
}

var updateGroupNameFactory = apiFactory[*UpdateGroupNameResponse, UpdateGroupNameFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UpdateGroupNameResponse]) (UpdateGroupNameFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/updateinfo", nil, true)

		return func(ctx context.Context, groupID string, name string) (*UpdateGroupNameResponse, error) {
			payload := map[string]any{
				"grid":  groupID,
				"gname": name,
				"imei":  sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateGroupName", err)
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

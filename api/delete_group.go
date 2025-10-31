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
	DeleteGroupResponse = string
	DeleteGroupFn       = func(ctx context.Context, groupID string) (DeleteGroupResponse, error)
)

func (a *api) DeleteGroup(ctx context.Context, groupID string) (DeleteGroupResponse, error) {
	return a.e.DeleteGroup(ctx, groupID)
}

var deleteGroupFactory = apiFactory[DeleteGroupResponse, DeleteGroupFn]()(
	func(a *api, sc session.Context, u factoryUtils[DeleteGroupResponse]) (DeleteGroupFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/disperse", nil, true)

		return func(ctx context.Context, groupID string) (DeleteGroupResponse, error) {
			payload := map[string]any{
				"grid": groupID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.DeleteGroup", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

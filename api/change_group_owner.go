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
	ChangeGroupOwnerResponse struct {
		Time int64 `json:"time"`
	}
	ChangeGroupOwnerFn = func(ctx context.Context, groupID string, memberID string) (*ChangeGroupOwnerResponse, error)
)

func (a *api) ChangeGroupOwner(ctx context.Context, groupID string, memberID string) (*ChangeGroupOwnerResponse, error) {
	return a.e.ChangeGroupOwner(ctx, groupID, memberID)
}

var changeGroupOwnerFactory = apiFactory[*ChangeGroupOwnerResponse, ChangeGroupOwnerFn]()(
	func(a *api, sc session.Context, u factoryUtils[*ChangeGroupOwnerResponse]) (ChangeGroupOwnerFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/change-owner", nil, true)

		return func(ctx context.Context, groupID string, memberID string) (*ChangeGroupOwnerResponse, error) {
			payload := map[string]any{
				"grid":       groupID,
				"newAdminId": memberID,
				"imei":       sc.IMEI(),
				"language":   sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ChangeGroupOwner", err)
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

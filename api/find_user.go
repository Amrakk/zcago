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

var ErrPhoneNumberEmpty = errs.NewZCA("phone number cannot be empty", "api.FindUser")

type (
	FindUserResponse map[string]struct {
		Avatar      string                 `json:"avatar"`
		Cover       string                 `json:"cover"`
		Status      string                 `json:"status"`
		Gender      model.Gender           `json:"gender"`
		DOB         int64                  `json:"dob"`
		Sdob        string                 `json:"sdob"`
		GlobalID    string                 `json:"globalId"`
		BizPkg      model.ZBusinessPackage `json:"bizPkg"`
		UID         string                 `json:"uid"`
		ZaloName    string                 `json:"zalo_name"`
		DisplayName string                 `json:"display_name"`
	}
	FindUserFn = func(ctx context.Context, phoneNumber ...string) (*FindUserResponse, error)
)

func (a *api) FindUser(ctx context.Context, phoneNumber ...string) (*FindUserResponse, error) {
	return a.e.FindUser(ctx, phoneNumber...)
}

var findUserFactory = apiFactory[*FindUserResponse, FindUserFn]()(
	func(a *api, sc session.Context, u factoryUtils[*FindUserResponse]) (FindUserFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/profile/multiget", nil, true)

		return func(ctx context.Context, phoneNumber ...string) (*FindUserResponse, error) {
			if len(phoneNumber) == 0 {
				return nil, ErrPhoneNumberEmpty
			}

			payload := map[string]any{
				"phones":      phoneNumber,
				"avatar_size": 240,
				"language":    sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.FindUser", err)
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

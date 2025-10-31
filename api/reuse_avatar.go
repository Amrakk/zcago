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
	ReuseAvatarResponse = struct{} // Always null
	ReuseAvatarFn       = func(ctx context.Context, photoID string) (*ReuseAvatarResponse, error)
)

func (a *api) ReuseAvatar(ctx context.Context, photoID string) (*ReuseAvatarResponse, error) {
	return a.e.ReuseAvatar(ctx, photoID)
}

var reuseAvatarFactory = apiFactory[*ReuseAvatarResponse, ReuseAvatarFn]()(
	func(a *api, sc session.Context, u factoryUtils[*ReuseAvatarResponse]) (ReuseAvatarFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/reuse-avatar", nil, true)

		return func(ctx context.Context, photoID string) (*ReuseAvatarResponse, error) {
			payload := map[string]any{
				"photoId":      photoID,
				"isPostSocial": 0,
				"imei":         sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ReuseAvatar", err)
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

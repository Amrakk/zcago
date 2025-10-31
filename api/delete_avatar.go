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
	DeleteAvatarResponse struct {
		DelPhotoIds []string     `json:"delPhotoIds"`
		ErrMap      model.ErrMap `json:"errMap"`
	}
	DeleteAvatarFn = func(ctx context.Context, photoID ...string) (*DeleteAvatarResponse, error)
)

func (a *api) DeleteAvatar(ctx context.Context, photoID ...string) (*DeleteAvatarResponse, error) {
	return a.e.DeleteAvatar(ctx, photoID...)
}

var deleteAvatarFactory = apiFactory[*DeleteAvatarResponse, DeleteAvatarFn]()(
	func(a *api, sc session.Context, u factoryUtils[*DeleteAvatarResponse]) (DeleteAvatarFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/del-avatars", nil, true)

		return func(ctx context.Context, photoID ...string) (*DeleteAvatarResponse, error) {
			delPhotos := make([]map[string]string, len(photoID))
			for i, id := range photoID {
				delPhotos[i] = map[string]string{"photoId": id}
			}

			payload := map[string]any{
				"delPhotos": jsonx.Stringify(delPhotos),
				"imei":      sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.DeleteAvatar", err)
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

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
	Photo struct {
		PhotoID   string `json:"photoId"`
		Thumbnail string `json:"thumbnail"`
		URL       string `json:"url"`
		BkURL     string `json:"bkUrl"`
	}
	GetAvatarListResponse struct {
		AlbumID     string  `json:"albumId"`
		NextPhotoID string  `json:"nextPhotoId"`
		HasMore     int     `json:"hasMore"`
		Photos      []Photo `json:"photos"`
	}
	GetAvatarListFn = func(ctx context.Context, options model.OffsetPaginationOptions) (*GetAvatarListResponse, error)
)

func (a *api) GetAvatarList(ctx context.Context, options model.OffsetPaginationOptions) (*GetAvatarListResponse, error) {
	return a.e.GetAvatarList(ctx, options)
}

var getAvatarListFactory = apiFactory[*GetAvatarListResponse, GetAvatarListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAvatarListResponse]) (GetAvatarListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/avatar-list", nil, true)

		return func(ctx context.Context, options model.OffsetPaginationOptions) (*GetAvatarListResponse, error) {
			if options.Count <= 0 {
				options.Count = 50
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			payload := map[string]any{
				"count":   options.Count,
				"page":    options.Page,
				"albumId": "0",
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetAvatarList", err)
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

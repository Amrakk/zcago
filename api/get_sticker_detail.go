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
	GetStickerDetailResponse = model.StickerDetail
	GetStickerDetailFn       = func(ctx context.Context, stickerID int) (*GetStickerDetailResponse, error)
)

func (a *api) GetStickerDetail(ctx context.Context, stickerID int) (*GetStickerDetailResponse, error) {
	return a.e.GetStickerDetail(ctx, stickerID)
}

var getStickerDetailFactory = apiFactory[*GetStickerDetailResponse, GetStickerDetailFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetStickerDetailResponse]) (GetStickerDetailFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("sticker"), "")
		serviceURL := u.MakeURL(base+"/api/message/sticker/sticker_detail", nil, true)

		return func(ctx context.Context, stickerID int) (*GetStickerDetailResponse, error) {
			payload := map[string]any{
				"sid": stickerID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetStickerDetail", err)
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

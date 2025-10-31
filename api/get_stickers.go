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
	GetStickersResponse = model.StickerSuggestions
	GetStickersFn       = func(ctx context.Context, keyword string) (*GetStickersResponse, error)
)

func (a *api) GetStickers(ctx context.Context, keyword string) (*GetStickersResponse, error) {
	return a.e.GetStickers(ctx, keyword)
}

var getStickersFactory = apiFactory[*GetStickersResponse, GetStickersFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetStickersResponse]) (GetStickersFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("sticker"), "")
		serviceURL := u.MakeURL(base+"/api/message/sticker/suggest/stickers", nil, true)

		return func(ctx context.Context, keyword string) (*GetStickersResponse, error) {
			payload := map[string]any{
				"keyword": keyword,
				"gif":     1,
				"guggy":   0,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetStickers", err)
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

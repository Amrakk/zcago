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
	GetSettingResponse = model.UserSetting
	GetSettingFn       = func(ctx context.Context) (*GetSettingResponse, error)
)

func (a *api) GetSetting(ctx context.Context) (*GetSettingResponse, error) {
	return a.e.GetSetting(ctx)
}

var getSettingFactory = apiFactory[*GetSettingResponse, GetSettingFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetSettingResponse]) (GetSettingFn, error) {
		serviceURL := u.MakeURL("https://wpa.chat.zalo.me/api/setting/me", nil, true)

		return func(ctx context.Context) (*GetSettingResponse, error) {
			payload := map[string]any{}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetSetting", err)
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

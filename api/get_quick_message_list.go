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
	GetQuickMessageListResponse struct {
		Cursor  int                  `json:"cursor"`
		Version int                  `json:"version"`
		Items   []model.QuickMessage `json:"items"`
	}
	GetQuickMessageListFn = func(ctx context.Context) (*GetQuickMessageListResponse, error)
)

func (a *api) GetQuickMessageList(ctx context.Context) (*GetQuickMessageListResponse, error) {
	return a.e.GetQuickMessageList(ctx)
}

var getQuickMessageListFactory = apiFactory[*GetQuickMessageListResponse, GetQuickMessageListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetQuickMessageListResponse]) (GetQuickMessageListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("quick_message"), "")
		serviceURL := u.MakeURL(base+"/api/quickmessage/list", nil, true)

		return func(ctx context.Context) (*GetQuickMessageListResponse, error) {
			payload := map[string]any{
				"version": 0,
				"lang":    0,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetQuickMessageList", err)
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

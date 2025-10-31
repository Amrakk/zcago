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
	HiddenThread struct {
		IsGroup  int    `json:"is_group"` // 1: true, 0: false
		ThreadID string `json:"thread_id"`
	}

	GetHiddenChatResponse struct {
		PIN     string         `json:"pin"`
		Threads []HiddenThread `json:"threads"`
	}
	GetHiddenChatFn = func(ctx context.Context) (*GetHiddenChatResponse, error)
)

func (a *api) GetHiddenChat(ctx context.Context) (*GetHiddenChatResponse, error) {
	return a.e.GetHiddenChat(ctx)
}

var getHiddenChatFactory = apiFactory[*GetHiddenChatResponse, GetHiddenChatFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetHiddenChatResponse]) (GetHiddenChatFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/hiddenconvers/get-all", nil, true)

		return func(ctx context.Context) (*GetHiddenChatResponse, error) {
			payload := map[string]any{
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetHiddenChat", err)
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

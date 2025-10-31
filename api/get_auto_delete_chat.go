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
	AutoDeleteConversation struct {
		DestID    string `json:"destId"`
		IsGroup   bool   `json:"isGroup"`
		TTL       int    `json:"ttl"`
		CreatedAt int64  `json:"createdAt"`
	}
	GetAutoDeleteChatResponse struct {
		Converts []AutoDeleteConversation `json:"converts"`
	}
	GetAutoDeleteChatFn = func(ctx context.Context) (*GetAutoDeleteChatResponse, error)
)

func (a *api) GetAutoDeleteChat(ctx context.Context) (*GetAutoDeleteChatResponse, error) {
	return a.e.GetAutoDeleteChat(ctx)
}

var getAutoDeleteChatFactory = apiFactory[*GetAutoDeleteChatResponse, GetAutoDeleteChatFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAutoDeleteChatResponse]) (GetAutoDeleteChatFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/conv/autodelete/getConvers", nil, true)

		return func(ctx context.Context) (*GetAutoDeleteChatResponse, error) {
			payload := map[string]any{}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetAutoDeleteChat", err)
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

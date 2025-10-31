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
	RemoveQuickMessageResponse struct {
		ItemIDs []int `json:"itemIds"`
		Version int   `json:"version"`
	}
	RemoveQuickMessageFn = func(ctx context.Context, messageID ...int) (*RemoveQuickMessageResponse, error)
)

func (a *api) RemoveQuickMessage(ctx context.Context, messageID ...int) (*RemoveQuickMessageResponse, error) {
	return a.e.RemoveQuickMessage(ctx, messageID...)
}

var removeQuickMessageFactory = apiFactory[*RemoveQuickMessageResponse, RemoveQuickMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*RemoveQuickMessageResponse]) (RemoveQuickMessageFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("quick_message"), "")
		serviceURL := u.MakeURL(base+"/api/quickmessage/delete", nil, true)

		return func(ctx context.Context, messageID ...int) (*RemoveQuickMessageResponse, error) {
			payload := map[string]any{
				"itemIds": messageID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.RemoveQuickMessage", err)
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

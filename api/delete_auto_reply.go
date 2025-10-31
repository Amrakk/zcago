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
	DeleteAutoReplyResponse struct {
		Item    int `json:"item"`
		Version int `json:"version"`
	}
	DeleteAutoReplyFn = func(ctx context.Context, id int) (*DeleteAutoReplyResponse, error)
)

func (a *api) DeleteAutoReply(ctx context.Context, id int) (*DeleteAutoReplyResponse, error) {
	return a.e.DeleteAutoReply(ctx, id)
}

var deleteAutoReplyFactory = apiFactory[*DeleteAutoReplyResponse, DeleteAutoReplyFn]()(
	func(a *api, sc session.Context, u factoryUtils[*DeleteAutoReplyResponse]) (DeleteAutoReplyFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("auto_reply"), "")
		serviceURL := u.MakeURL(base+"/api/autoreply/delete", nil, true)

		return func(ctx context.Context, id int) (*DeleteAutoReplyResponse, error) {
			payload := map[string]any{
				"id":      id,
				"cliLang": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.DeleteAutoReply", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

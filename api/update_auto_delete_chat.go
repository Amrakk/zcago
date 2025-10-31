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

type ChatTTL int

const (
	ChatTTLOff ChatTTL = 0
	ChatTTL24H ChatTTL = 86400000
	ChatTTL7D  ChatTTL = 7 * ChatTTL24H
	ChatTTL14D ChatTTL = 2 * ChatTTL7D
)

type (
	UpdateAutoDeleteChatResponse = string
	UpdateAutoDeleteChatFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, ttl ChatTTL) (UpdateAutoDeleteChatResponse, error)
)

func (a *api) UpdateAutoDeleteChat(ctx context.Context, threadID string, threadType model.ThreadType, ttl ChatTTL) (UpdateAutoDeleteChatResponse, error) {
	return a.e.UpdateAutoDeleteChat(ctx, threadID, threadType, ttl)
}

var updateAutoDeleteChatFactory = apiFactory[UpdateAutoDeleteChatResponse, UpdateAutoDeleteChatFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateAutoDeleteChatResponse]) (UpdateAutoDeleteChatFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/conv/autodelete/updateConvers", nil, true)

		return func(ctx context.Context, threadID string, threadType model.ThreadType, ttl ChatTTL) (UpdateAutoDeleteChatResponse, error) {
			payload := map[string]any{
				"threadId":   threadID,
				"isGroup":    jsonx.B2I(threadType == model.ThreadTypeGroup),
				"ttl":        ttl,
				"clientLang": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateAutoDeleteChat", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

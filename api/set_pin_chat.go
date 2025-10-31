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
	SetPinChatResponse = string
	SetPinChatFn       = func(ctx context.Context, threadID []string, threadType model.ThreadType, isPinned bool) (SetPinChatResponse, error)
)

func (a *api) SetPinChat(ctx context.Context, threadID []string, threadType model.ThreadType, isPinned bool) (SetPinChatResponse, error) {
	return a.e.SetPinChat(ctx, threadID, threadType, isPinned)
}

var setPinChatFactory = apiFactory[SetPinChatResponse, SetPinChatFn]()(
	func(a *api, sc session.Context, u factoryUtils[SetPinChatResponse]) (SetPinChatFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/pinconvers/updatev2", nil, true)

		return func(ctx context.Context, threadID []string, threadType model.ThreadType, isPinned bool) (SetPinChatResponse, error) {
			prefix := "g"
			if threadType == model.ThreadTypeUser {
				prefix = "u"
			}

			conversations := make([]string, len(threadID))
			for i, id := range threadID {
				conversations[i] = prefix + id
			}

			payload := map[string]any{
				"actionType":    jsonx.B2I(!isPinned) + 1,
				"conversations": conversations,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SetPinChat", err)
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

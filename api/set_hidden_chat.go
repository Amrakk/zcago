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
	SetHiddenChatResponse = string
	SetHiddenChatFn       = func(ctx context.Context, threadID []string, threadType model.ThreadType, isHidden bool) (SetHiddenChatResponse, error)
)

func (a *api) SetHiddenChat(ctx context.Context, threadID []string, threadType model.ThreadType, isHidden bool) (SetHiddenChatResponse, error) {
	return a.e.SetHiddenChat(ctx, threadID, threadType, isHidden)
}

var setHiddenChatFactory = apiFactory[SetHiddenChatResponse, SetHiddenChatFn]()(
	func(a *api, sc session.Context, u factoryUtils[SetHiddenChatResponse]) (SetHiddenChatFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/hiddenconvers/add-remove", nil, true)

		return func(ctx context.Context, threadID []string, threadType model.ThreadType, isHidden bool) (SetHiddenChatResponse, error) {
			activeKey := "add_threads"
			inactiveKey := "del_threads"
			if !isHidden {
				activeKey, inactiveKey = inactiveKey, activeKey
			}

			isGroup := jsonx.B2I(threadType == model.ThreadTypeGroup)
			threadIDs := make([]any, len(threadID))
			for i, id := range threadID {
				threadIDs[i] = map[string]any{
					"thread_id": id,
					"is_group":  isGroup,
				}
			}

			payload := map[string]any{
				activeKey:   jsonx.Stringify(threadIDs),
				inactiveKey: "[]",
				"imei":      sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SetHiddenChat", err)
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

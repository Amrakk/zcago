package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	DeleteChatLastMessage struct {
		OwnerID     string
		CliMsgID    string
		GlobalMsgID string
	}
	DeleteChatResponse struct {
		Status int `json:"status"`
	}
	DeleteChatFn = func(ctx context.Context, threadID string, threadType model.ThreadType, lastMsg DeleteChatLastMessage) (*DeleteChatResponse, error)
)

func (a *api) DeleteChat(ctx context.Context, threadID string, threadType model.ThreadType, lastMsg DeleteChatLastMessage) (*DeleteChatResponse, error) {
	return a.e.DeleteChat(ctx, threadID, threadType, lastMsg)
}

var deleteChatFactory = apiFactory[*DeleteChatResponse, DeleteChatFn]()(
	func(a *api, sc session.Context, u factoryUtils[*DeleteChatResponse]) (DeleteChatFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")
		defaultParams := map[string]any{"nretry": 0}

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/deleteconver", defaultParams, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/deleteconver", defaultParams, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, lastMsg DeleteChatLastMessage) (*DeleteChatResponse, error) {
			key := "grid"
			if threadType == model.ThreadTypeUser {
				key = "toid"
			}

			conver := map[string]any{
				"ownerId":     lastMsg.OwnerID,
				"cliMsgId":    lastMsg.CliMsgID,
				"globalMsgId": lastMsg.GlobalMsgID,
			}

			payload := map[string]any{
				key:        threadID,
				"cliMsgId": strconv.FormatInt(time.Now().UnixMilli(), 10),
				"conver":   conver,
				"onlyMe":   1,
				"imei":     sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.DeleteChat", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[threadType], &httpx.RequestOptions{
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

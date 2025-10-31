package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	UndoMessageData struct {
		MsgID    string
		CliMsgID string
	}
	UndoMessageResponse struct {
		Status int `json:"status"`
	}
	UndoMessageFn = func(ctx context.Context, threadID string, threadType model.ThreadType, data UndoMessageData) (*UndoMessageResponse, error)
)

func (a *api) UndoMessage(ctx context.Context, threadID string, threadType model.ThreadType, data UndoMessageData) (*UndoMessageResponse, error) {
	return a.e.UndoMessage(ctx, threadID, threadType, data)
}

var undoMessageFactory = apiFactory[*UndoMessageResponse, UndoMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UndoMessageResponse]) (UndoMessageFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/undo", nil, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/undomsg", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, data UndoMessageData) (*UndoMessageResponse, error) {
			isDM := threadType == model.ThreadTypeUser

			key := "grid"
			if isDM {
				key = "toid"
			}

			payload := map[string]any{
				key:            threadID,
				"msgId":        data.MsgID,
				"cliMsgIdUndo": data.CliMsgID,
				"clientId":     time.Now().UnixMilli(),
			}

			if !isDM {
				payload["visibility"] = 0
				payload["imei"] = sc.IMEI()
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UndoMessage", err)
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

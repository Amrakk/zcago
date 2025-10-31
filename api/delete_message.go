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

var (
	ErrSenderRecall = errs.NewZCA(
		"to delete a message you sent from others' view, call UndoMessage instead",
		"api.DeleteMessage",
	)
	ErrDMRecipientsDeleteUnsupported = errs.NewZCA(
		"delete for everyone is not supported in direct messages",
		"api.DeleteMessage",
	)
)

type (
	DeleteMessageData struct {
		CliMsgID string
		MsgID    string
		UIDFrom  string
	}
	DeleteMessageDestination struct {
		ThreadID string
		Type     model.ThreadType
		Data     DeleteMessageData
	}
	DeleteMessageResponse struct {
		Status int `json:"status"`
	}
	DeleteMessageFn = func(ctx context.Context, dest DeleteMessageDestination, onlyMe bool) (*DeleteMessageResponse, error)
)

func (a *api) DeleteMessage(ctx context.Context, dest DeleteMessageDestination, onlyMe bool) (*DeleteMessageResponse, error) {
	return a.e.DeleteMessage(ctx, dest, onlyMe)
}

var deleteMessageFactory = apiFactory[*DeleteMessageResponse, DeleteMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*DeleteMessageResponse]) (DeleteMessageFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")
		defaultParams := map[string]any{"nretry": 0}

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/delete", defaultParams, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/deletemsg", defaultParams, true),
		}

		return func(ctx context.Context, dest DeleteMessageDestination, onlyMe bool) (*DeleteMessageResponse, error) {
			threadID, threadType, data := dest.ThreadID, dest.Type, dest.Data
			isDM := threadType == model.ThreadTypeUser

			if sc.UID() == data.UIDFrom && !onlyMe {
				return nil, ErrSenderRecall
			}
			if isDM && !onlyMe {
				return nil, ErrDMRecipientsDeleteUnsupported
			}

			key := "grid"
			if isDM {
				key = "toid"
			}

			payload := map[string]any{
				key:        threadID,
				"cliMsgId": time.Now().UnixMilli(),
				"msgs": []map[string]any{
					{
						"ownerId":     data.UIDFrom,
						"cliMsgId":    data.CliMsgID,
						"globalMsgId": data.MsgID,
						"destId":      threadID,
					},
				},
				"onlyMe": jsonx.B2I(onlyMe),
			}

			if isDM {
				payload["imei"] = sc.IMEI()
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.DeleteMessage", err)
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

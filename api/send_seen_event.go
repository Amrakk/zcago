package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/config"
	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	SendSeenEventResponse struct {
		Status int `json:"status"`
	}
	SendSeenEventFn = func(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType) (*SendSeenEventResponse, error)
)

func (a *api) SendSeenEvent(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType) (*SendSeenEventResponse, error) {
	return a.e.SendSeenEvent(ctx, messages, threadType)
}

var sendSeenEventFactory = apiFactory[*SendSeenEventResponse, SendSeenEventFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendSeenEventResponse]) (SendSeenEventFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")
		defaultParams := map[string]any{"nretry": 0}

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/seenv2", defaultParams, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/seenv2", defaultParams, true),
		}

		return func(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType) (*SendSeenEventResponse, error) {
			if n := len(messages); n == 0 || n > config.MaxMessagesPerRequest {
				return nil, errs.ErrInvalidMessageCount
			}

			isGroup := threadType == model.ThreadTypeGroup

			key := "senderId"
			threadID := messages[0].UIDFrom
			if isGroup {
				key = "grid"
				threadID = messages[0].IDTo
			}

			data := make([]map[string]any, len(messages))
			for i, msg := range messages {
				currThreadID := msg.UIDFrom
				if isGroup {
					currThreadID = msg.IDTo
				}
				if currThreadID != threadID {
					return nil, errs.ErrInconsistentGroupRecipient
				}

				di := "0"
				if msg.IDTo != sc.UID() {
					di = msg.IDTo
				}

				data[i] = map[string]any{
					"cmi": msg.CliMsgID,
					"gmi": msg.MsgID,
					"si":  msg.UIDFrom,
					"di":  di,
					"mt":  msg.MsgType,
					"st":  msg.ST,
					"at":  msg.AT,
					"cmd": msg.CMD,
					"ts":  msg.TS,
				}
			}

			msgInfos := map[string]any{
				key:    threadID,
				"data": data,
			}

			payload := map[string]any{
				"msgInfos": jsonx.Stringify(msgInfos),
			}

			if isGroup {
				payload["imei"] = sc.IMEI()
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendSeenEvent", err)
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

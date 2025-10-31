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
	SendDeliveredEventResponse = string // string | { status: int }
	SendDeliveredEventFn       = func(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType, isSeen bool) (SendDeliveredEventResponse, error)
)

func (a *api) SendDeliveredEvent(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType, isSeen bool) (SendDeliveredEventResponse, error) {
	return a.e.SendDeliveredEvent(ctx, messages, threadType, isSeen)
}

var sendDeliveredEventFactory = apiFactory[SendDeliveredEventResponse, SendDeliveredEventFn]()(
	func(a *api, sc session.Context, u factoryUtils[SendDeliveredEventResponse]) (SendDeliveredEventFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/deliveredv2", nil, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/deliveredv2", nil, true),
		}

		return func(ctx context.Context, messages []model.OutboundMessage, threadType model.ThreadType, isSeen bool) (SendDeliveredEventResponse, error) {
			if n := len(messages); n == 0 || n > config.MaxMessagesPerRequest {
				return "", errs.ErrInvalidMessageCount
			}

			isGroup := threadType == model.ThreadTypeGroup

			// 27/02/2025
			// This can send messages from multiple groups, but to prevent potential issues,
			// we will restrict it to sending messages only within the same group.
			idTo := messages[0].IDTo
			if isGroup {
				for _, msg := range messages {
					if msg.IDTo != idTo {
						return "", errs.ErrInconsistentGroupRecipient
					}
				}
			}

			di := "0"
			if idTo != sc.UID() {
				di = idTo
			}

			data := make([]map[string]any, len(messages))
			for i, msg := range messages {
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

			payload := map[string]any{}
			msgInfos := map[string]any{
				"seen": jsonx.B2I(isSeen),
				"data": data,
			}

			if threadType == model.ThreadTypeUser {
				msgInfos["grid"] = idTo
				payload["imei"] = sc.IMEI()
			}

			payload["msgInfos"] = jsonx.Stringify(msgInfos)

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SendDeliveredEvent", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[threadType], &httpx.RequestOptions{
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

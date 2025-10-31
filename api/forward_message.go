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

var (
	ErrMessageEmpty  = errs.NewZCA("message content cannot be empty", "api.ForwardMessage")
	ErrThreadIDEmpty = errs.NewZCA("threadID cannot be empty", "api.ForwardMessage")
)

type (
	ForwardMessageReference struct {
		ID         string
		TS         int
		LogSrcType int
		FwLvl      int
	}
	ForwardMessagePayload struct {
		Message   string
		TTL       int
		Reference *ForwardMessageReference
	}

	ForwardMessageSuccess struct {
		ClientID string `json:"clientId"`
		MsgID    string `json:"msgId"`
	}
	ForwardMessageFail struct {
		ClientID  string `json:"clientId"`
		ErrorCode string `json:"error_code"`
	}

	ForwardMessageResponse struct {
		Success []ForwardMessageSuccess `json:"success"`
		Fail    []ForwardMessageFail    `json:"fail"`
	}
	ForwardMessageFn = func(ctx context.Context, threadIDs []string, threadType model.ThreadType, message ForwardMessagePayload) (*ForwardMessageResponse, error)
)

func (a *api) ForwardMessage(ctx context.Context, threadIDs []string, threadType model.ThreadType, message ForwardMessagePayload) (*ForwardMessageResponse, error) {
	return a.e.ForwardMessage(ctx, threadIDs, threadType, message)
}

var forwardMessageFactory = apiFactory[*ForwardMessageResponse, ForwardMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*ForwardMessageResponse]) (ForwardMessageFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/mforward", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/mforward", nil, true),
		}

		return func(ctx context.Context, threadIDs []string, threadType model.ThreadType, message ForwardMessagePayload) (*ForwardMessageResponse, error) {
			if len(message.Message) == 0 {
				return nil, ErrMessageEmpty
			}
			if len(threadIDs) == 0 {
				return nil, ErrThreadIDEmpty
			}

			clientID := strconv.FormatInt(time.Now().UnixMilli(), 10)
			msgInfo := map[string]any{"message": message.Message}

			var decorLog any
			if ref := message.Reference; ref != nil {
				msgInfo["reference"] = jsonx.Stringify(map[string]any{
					"type": 3,
					"data": jsonx.Stringify(ref),
				})

				fw := map[string]any{
					"pmsg":  map[string]any{"st": 1, "ts": ref.TS, "id": ref.ID},
					"rmsg":  map[string]any{"st": 1, "ts": ref.TS, "id": ref.ID},
					"fwLvl": ref.FwLvl,
				}
				decorLog = map[string]any{"fw": fw}
			}

			key := "grid"
			if threadType == model.ThreadTypeUser {
				key = "toUid"
			}

			recipients := make([]map[string]any, len(threadIDs))
			for i, id := range threadIDs {
				recipients[i] = map[string]any{
					key:        id,
					"clientId": clientID,
					"ttl":      message.TTL,
				}
			}

			payload := map[string]any{
				"ttl":      message.TTL,
				"msgType":  "1",
				"totalIds": len(threadIDs),
				"msgInfo":  jsonx.Stringify(msgInfo),
				"decorLog": jsonx.Stringify(decorLog),
			}

			if threadType == model.ThreadTypeUser {
				payload["toIds"] = recipients
				payload["imei"] = sc.IMEI()
			} else {
				payload["grids"] = recipients
			}

			// raw, _ := json.MarshalIndent(params, "", "  ")
			// fmt.Println(string(raw))
			// fmt.Println(serviceURLs[threadType])

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ForwardMessage", err)
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

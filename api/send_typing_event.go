package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	SendTypingEventResponse struct {
		Status int `json:"status"`
	}
	SendTypingEventFn = func(ctx context.Context, threadID string, threadType model.ThreadType, destType model.DestType) (*SendTypingEventResponse, error)
)

func (a *api) SendTypingEvent(ctx context.Context, threadID string, threadType model.ThreadType, destType model.DestType) (*SendTypingEventResponse, error) {
	return a.e.SendTypingEvent(ctx, threadID, threadType, destType)
}

var sendTypingEventFactory = apiFactory[*SendTypingEventResponse, SendTypingEventFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendTypingEventResponse]) (SendTypingEventFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/typing", nil, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/typing", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, destType model.DestType) (*SendTypingEventResponse, error) {
			payload := map[string]any{
				"imei": sc.IMEI(),
			}

			if threadType == model.ThreadTypeUser {
				payload["toid"] = threadID
				payload["destType"] = destType
			} else {
				payload["grid"] = threadID
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendTypingEvent", err)
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

func (r *SendTypingEventResponse) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		*r = SendTypingEventResponse{}
		return nil
	}

	type alias SendTypingEventResponse
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*r = SendTypingEventResponse(tmp)

	return nil
}

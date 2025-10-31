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
	SendVoiceOptions struct {
		VoiceURL string // URL of the voice
		TTL      int    // Time to live in milliseconds
	}
	SendVoiceResponse struct {
		MsgID int `json:"msgId"`
	}
	SendVoiceFn = func(ctx context.Context, threadID string, threadType model.ThreadType, options SendVoiceOptions) (*SendVoiceResponse, error)
)

func (a *api) SendVoice(ctx context.Context, threadID string, threadType model.ThreadType, options SendVoiceOptions) (*SendVoiceResponse, error) {
	return a.e.SendVoice(ctx, threadID, threadType, options)
}

var sendVoiceFactory = apiFactory[*SendVoiceResponse, SendVoiceFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendVoiceResponse]) (SendVoiceFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/forward", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/forward", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options SendVoiceOptions) (*SendVoiceResponse, error) {
			head, err := u.Request(ctx, options.VoiceURL, &httpx.RequestOptions{
				Method: http.MethodHead,
				Raw:    true,
			})
			if err != nil {
				return nil, errs.ErrFileContentUnavailable
			}
			defer head.Body.Close()

			fileSize := head.ContentLength
			if fileSize == -1 {
				fileSize = 0
			}

			msgInfo := map[string]any{
				"voiceUrl": options.VoiceURL,
				"m4aUrl":   options.VoiceURL,
				"fileSize": fileSize,
			}

			payload := map[string]any{
				"clientId": strconv.FormatInt(time.Now().UnixMilli(), 10),
				"ttl":      options.TTL,
				"zsource":  -1,
				"msgType":  3,
				"msgInfo":  jsonx.Stringify(msgInfo),
				"imei":     sc.IMEI(),
			}

			if threadType == model.ThreadTypeUser {
				payload["toid"] = threadID
			} else {
				payload["grid"] = threadID
				payload["visibility"] = 0
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendVoice", err)
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

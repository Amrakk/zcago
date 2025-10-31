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
	SendStickerPayload struct {
		ID     int `json:"id"`
		CateID int `json:"cateId"`
		Type   int `json:"type"`
	}
	SendStickerResponse struct {
		MsgID string `json:"msgId"`
	}
	SendStickerFn = func(ctx context.Context, threadID string, threadType model.ThreadType, sticker SendStickerPayload) (*SendStickerResponse, error)
)

func (a *api) SendSticker(ctx context.Context, threadID string, threadType model.ThreadType, sticker SendStickerPayload) (*SendStickerResponse, error) {
	return a.e.SendSticker(ctx, threadID, threadType, sticker)
}

var sendStickerFactory = apiFactory[*SendStickerResponse, SendStickerFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendStickerResponse]) (SendStickerFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")
		defaultParams := map[string]any{"nretry": 0}

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/sticker", defaultParams, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/sticker", defaultParams, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, sticker SendStickerPayload) (*SendStickerResponse, error) {
			key := "grid"
			if threadType == model.ThreadTypeUser {
				key = "toid"
			}

			payload := map[string]any{
				key:         threadID,
				"stickerId": sticker.ID,
				"cateId":    sticker.CateID,
				"type":      sticker.Type,
				"clientId":  time.Now().UnixMilli(),
				"imei":      sc.IMEI(),
				"zsource":   101,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendSticker", err)
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

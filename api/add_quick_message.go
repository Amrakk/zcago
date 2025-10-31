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
	AddQuickMessageRequest struct {
		Keyword string
		Title   string
		Media   *model.AttachmentSource
	}
	AddQuickMessageResponse struct {
		Item    model.QuickMessage `json:"item"`
		Version int                `json:"version"`
	}
	AddQuickMessageFn = func(ctx context.Context, message AddQuickMessageRequest) (*AddQuickMessageResponse, error)
)

func (a *api) AddQuickMessage(ctx context.Context, message AddQuickMessageRequest) (*AddQuickMessageResponse, error) {
	return a.e.AddQuickMessage(ctx, message)
}

var addQuickMessageFactory = apiFactory[*AddQuickMessageResponse, AddQuickMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*AddQuickMessageResponse]) (AddQuickMessageFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("quick_message"), "")
		serviceURL := u.MakeURL(base+"/api/quickmessage/create", nil, true)

		return func(ctx context.Context, message AddQuickMessageRequest) (*AddQuickMessageResponse, error) {
			msgType := model.QuickMessageTypeText
			if message.Media != nil {
				msgType = model.QuickMessageTypeMedia
			}

			payload := map[string]any{
				"keyword": message.Keyword,
				"message": map[string]any{
					"title":  message.Title,
					"params": "",
				},
				"type": msgType,
				"imei": sc.IMEI(),
			}

			if msgType == model.QuickMessageTypeMedia {
				up, err := a.UploadPhoto(ctx, *message.Media)
				if err != nil {
					return nil, errs.WrapZCA("failed to upload media", "api.AddQuickMessage", err)
				}

				payload["media"] = map[string]any{
					"items": []any{
						map[string]any{
							"type":         0,
							"photoId":      up.PhotoID,
							"title":        "",
							"width":        "",
							"height":       "",
							"previewThumb": up.ThumbURL,
							"rawUrl":       jsonx.Or(up.NormalURL, up.HdURL),
							"thumbUrl":     up.ThumbURL,
							"normalUrl":    jsonx.Or(up.NormalURL, up.HdURL),
							"hdUrl":        jsonx.Or(up.HdURL, up.NormalURL),
						},
					},
				}
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.AddQuickMessage", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

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
	UpdateQuickMessageData struct {
		Keyword string
		Title   string
		Media   *model.AttachmentSource
	}
	UpdateQuickMessageResponse struct {
		Item    model.QuickMessage `json:"item"`
		Version int                `json:"version"`
	}
	UpdateQuickMessageFn = func(ctx context.Context, itemID string, data UpdateQuickMessageData) (*UpdateQuickMessageResponse, error)
)

func (a *api) UpdateQuickMessage(ctx context.Context, itemID string, data UpdateQuickMessageData) (*UpdateQuickMessageResponse, error) {
	return a.e.UpdateQuickMessage(ctx, itemID, data)
}

var updateQuickMessageFactory = apiFactory[*UpdateQuickMessageResponse, UpdateQuickMessageFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UpdateQuickMessageResponse]) (UpdateQuickMessageFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("quick_message"), "")
		serviceURL := u.MakeURL(base+"/api/quickmessage/update", nil, true)

		return func(ctx context.Context, itemID string, data UpdateQuickMessageData) (*UpdateQuickMessageResponse, error) {
			msgType := model.QuickMessageTypeText
			if data.Media != nil {
				msgType = model.QuickMessageTypeMedia
			}

			payload := map[string]any{
				"itemId":  itemID,
				"keyword": data.Keyword,
				"message": map[string]any{
					"title":  data.Title,
					"params": "",
				},
				"type": msgType,
			}

			if msgType == model.QuickMessageTypeMedia {
				up, err := a.UploadPhoto(ctx, *data.Media)
				if err != nil {
					return nil, errs.WrapZCA("failed to upload media", "api.UpdateQuickMessage", err)
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
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateQuickMessage", err)
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

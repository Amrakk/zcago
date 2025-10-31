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
	SendVideoOptions struct {
		Msg          string // Optional message to send along with the video
		VideoURL     string // URL of the video
		ThumbnailURL string // URL of the thumbnail
		Duration     int    // Video duration in milliseconds
		Width        int    // Width of the video
		Height       int    // Height of the video
		TTL          int    // Time to live in milliseconds
	}
	SendVideoResponse struct {
		MsgID int `json:"msgId"`
	}
	SendVideoFn = func(ctx context.Context, threadID string, threadType model.ThreadType, options SendVideoOptions) (*SendVideoResponse, error)
)

func (a *api) SendVideo(ctx context.Context, threadID string, threadType model.ThreadType, options SendVideoOptions) (*SendVideoResponse, error) {
	return a.e.SendVideo(ctx, threadID, threadType, options)
}

var sendVideoFactory = apiFactory[*SendVideoResponse, SendVideoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendVideoResponse]) (SendVideoFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/forward", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/forward", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options SendVideoOptions) (*SendVideoResponse, error) {
			head, err := u.Request(ctx, options.VideoURL, &httpx.RequestOptions{
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

			width := 1280
			if options.Width != 0 {
				width = options.Width
			}

			height := 720
			if options.Height != 0 {
				height = options.Height
			}

			msgInfo := map[string]any{
				"videoUrl": options.VideoURL,
				"thumbUrl": options.ThumbnailURL,
				"duration": options.Duration,
				"width":    width,
				"height":   height,
				"fileSize": fileSize,
				"properties": map[string]any{
					"color":   -1,
					"size":    -1,
					"type":    1003,
					"subType": 0,
					"ext": map[string]any{
						"sSrcType":         -1,
						"sSrcStr":          "",
						"msg_warning_type": 0,
					},
				},
				"title": options.Msg,
			}

			payload := map[string]any{
				"clientId": strconv.FormatInt(time.Now().UnixMilli(), 10),
				"ttl":      options.TTL,
				"zsource":  704,
				"msgType":  5,
				"msgInfo":  jsonx.Stringify(msgInfo),
				"imei":     sc.IMEI(),
			}

			if threadType == model.ThreadTypeUser {
				payload["toId"] = threadID
			} else {
				payload["grid"] = threadID
				payload["visibility"] = 0
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendVideo", err)
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

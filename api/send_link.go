package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	SendLinkOptions struct {
		Msg      string
		Link     string
		TTL      int
		Mentions []model.TMention
	}
	SendLinkResponse struct {
		MsgID string `json:"msgId"`
	}
	SendLinkFn = func(ctx context.Context, threadID string, threadType model.ThreadType, options SendLinkOptions) (*SendLinkResponse, error)
)

func (a *api) SendLink(ctx context.Context, threadID string, threadType model.ThreadType, options SendLinkOptions) (*SendLinkResponse, error) {
	return a.e.SendLink(ctx, threadID, threadType, options)
}

var sendLinkFactory = apiFactory[*SendLinkResponse, SendLinkFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendLinkResponse]) (SendLinkFn, error) {
		userBase := jsonx.FirstOr(sc.GetZpwService("chat"), "")
		groupBase := jsonx.FirstOr(sc.GetZpwService("group"), "")
		defaultParams := map[string]any{"nretry": 0}

		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(userBase+"/api/message/link", defaultParams, true),
			model.ThreadTypeGroup: u.MakeURL(groupBase+"/api/group/sendlink", defaultParams, true),
		}
		return func(ctx context.Context, threadID string, threadType model.ThreadType, options SendLinkOptions) (*SendLinkResponse, error) {
			linkData, err := a.ParseLink(ctx, options.Link)
			if err != nil {
				return nil, err
			}

			isGroup := threadType == model.ThreadTypeGroup

			msg := options.Link
			if s := strings.TrimSpace(options.Msg); s != "" {
				if strings.Contains(s, options.Link) {
					msg = s
				} else {
					msg = s + " " + options.Link
				}
			}

			d := linkData.Data

			payload := map[string]any{
				"msg":      msg,
				"href":     d.Href,
				"src":      d.Src,
				"title":    d.Title,
				"desc":     d.Desc,
				"thumb":    d.Thumb,
				"type":     2,
				"media":    jsonx.Stringify(d.Media),
				"ttl":      options.TTL,
				"clientId": time.Now().UnixMilli(),
			}

			if isGroup {
				payload["grid"] = threadID
				payload["imei"] = sc.IMEI()
				payload["mentionInfo"] = jsonx.Stringify(options.Mentions)
			} else {
				payload["toId"] = threadID
				payload["mentionInfo"] = ""
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendLink", err)
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

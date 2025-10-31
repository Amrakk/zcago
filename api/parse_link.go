package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	ParseLinkMedia struct {
		Type       int    `json:"type"`
		Count      int    `json:"count"`
		MediaTitle string `json:"mediaTitle"`
		Artist     string `json:"artist"`
		StreamURL  string `json:"streamUrl"`
		StreamIcon string `json:"stream_icon"`
	}
	ParseLinkData struct {
		Thumb      string         `json:"thumb"`
		Title      string         `json:"title"`
		Desc       string         `json:"desc"`
		Src        string         `json:"src"`
		Href       string         `json:"href"`
		Media      ParseLinkMedia `json:"media"`
		StreamIcon string         `json:"stream_icon"`
	}

	ParseLinkResponse struct {
		Data     ParseLinkData  `json:"data"`
		ErrorMap map[string]int `json:"error_maps"`
		// ErrorMap model.ErrMap `json:"error_maps"`
	}
	ParseLinkFn = func(ctx context.Context, link string) (*ParseLinkResponse, error)
)

func (a *api) ParseLink(ctx context.Context, link string) (*ParseLinkResponse, error) {
	return a.e.ParseLink(ctx, link)
}

var parseLinkFactory = apiFactory[*ParseLinkResponse, ParseLinkFn]()(
	func(a *api, sc session.Context, u factoryUtils[*ParseLinkResponse]) (ParseLinkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURL := u.MakeURL(base+"/api/message/parselink", nil, true)

		return func(ctx context.Context, link string) (*ParseLinkResponse, error) {
			payload := map[string]any{
				"link":    link,
				"version": 1,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ParseLink", err)
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

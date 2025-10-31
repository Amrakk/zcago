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
	JoinGroupLinkResponse = string
	JoinGroupLinkFn       = func(ctx context.Context, link string) (JoinGroupLinkResponse, error)
)

func (a *api) JoinGroupLink(ctx context.Context, link string) (JoinGroupLinkResponse, error) {
	return a.e.JoinGroupLink(ctx, link)
}

var joinGroupLinkFactory = apiFactory[JoinGroupLinkResponse, JoinGroupLinkFn]()(
	func(a *api, sc session.Context, u factoryUtils[JoinGroupLinkResponse]) (JoinGroupLinkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/link/join", nil, true)

		return func(ctx context.Context, link string) (JoinGroupLinkResponse, error) {
			payload := map[string]any{
				"link":       link,
				"clientLang": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.JoinGroupLink", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

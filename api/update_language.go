package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type Language string

const (
	LanguageVietnamese Language = "VI"
	LanguageEnglish    Language = "EN"
)

type (
	UpdateLanguageResponse = string
	UpdateLanguageFn       = func(ctx context.Context, lang Language) (UpdateLanguageResponse, error)
)

func (a *api) UpdateLanguage(ctx context.Context, lang Language) (UpdateLanguageResponse, error) {
	return a.e.UpdateLanguage(ctx, lang)
}

var updateLanguageFactory = apiFactory[UpdateLanguageResponse, UpdateLanguageFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateLanguageResponse]) (UpdateLanguageFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/profile/updatelang", nil, true)

		return func(ctx context.Context, lang Language) (UpdateLanguageResponse, error) {
			payload := map[string]any{
				"language": string(lang),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateLanguage", err)
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

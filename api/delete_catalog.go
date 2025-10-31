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
	DeleteCatalogResponse = string
	DeleteCatalogFn       = func(ctx context.Context, catalogID string) (DeleteCatalogResponse, error)
)

func (a *api) DeleteCatalog(ctx context.Context, catalogID string) (DeleteCatalogResponse, error) {
	return a.e.DeleteCatalog(ctx, catalogID)
}

var deleteCatalogFactory = apiFactory[DeleteCatalogResponse, DeleteCatalogFn]()(
	func(a *api, sc session.Context, u factoryUtils[DeleteCatalogResponse]) (DeleteCatalogFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("catalog"), "")
		serviceURL := u.MakeURL(base+"/api/prodcatalog/catalog/delete", nil, true)

		return func(ctx context.Context, catalogID string) (DeleteCatalogResponse, error) {
			payload := map[string]any{
				"catalog_id": catalogID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.DeleteCatalog", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

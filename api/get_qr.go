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
	GetQRResponse map[string]string
	GetQRFn       = func(ctx context.Context, userID ...string) (*GetQRResponse, error)
)

func (a *api) GetQR(ctx context.Context, userID ...string) (*GetQRResponse, error) {
	return a.e.GetQR(ctx, userID...)
}

var getQRFactory = apiFactory[*GetQRResponse, GetQRFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetQRResponse]) (GetQRFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/mget-qr", nil, true)

		return func(ctx context.Context, userID ...string) (*GetQRResponse, error) {
			payload := map[string]any{
				"fids": userID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetQR", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
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

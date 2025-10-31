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
	SharePollResponse = string
	SharePollFn       = func(ctx context.Context, pollID int) (SharePollResponse, error)
)

func (a *api) SharePoll(ctx context.Context, pollID int) (SharePollResponse, error) {
	return a.e.SharePoll(ctx, pollID)
}

var sharePollFactory = apiFactory[SharePollResponse, SharePollFn]()(
	func(a *api, sc session.Context, u factoryUtils[SharePollResponse]) (SharePollFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/poll/share", nil, true)

		return func(ctx context.Context, pollID int) (SharePollResponse, error) {
			payload := map[string]any{
				"poll_id": pollID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SharePoll", err)
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

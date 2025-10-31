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
	LockPollResponse = string
	LockPollFn       = func(ctx context.Context, pollID int) (LockPollResponse, error)
)

func (a *api) LockPoll(ctx context.Context, pollID int) (LockPollResponse, error) {
	return a.e.LockPoll(ctx, pollID)
}

var lockPollFactory = apiFactory[LockPollResponse, LockPollFn]()(
	func(a *api, sc session.Context, u factoryUtils[LockPollResponse]) (LockPollFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/poll/end", nil, true)

		return func(ctx context.Context, pollID int) (LockPollResponse, error) {
			payload := map[string]any{
				"poll_id": pollID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.LockPoll", err)
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

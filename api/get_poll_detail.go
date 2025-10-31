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
	GetPollDetailResponse = model.PollDetail
	GetPollDetailFn       = func(ctx context.Context, pollID int) (*GetPollDetailResponse, error)
)

func (a *api) GetPollDetail(ctx context.Context, pollID int) (*GetPollDetailResponse, error) {
	return a.e.GetPollDetail(ctx, pollID)
}

var getPollDetailFactory = apiFactory[*GetPollDetailResponse, GetPollDetailFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetPollDetailResponse]) (GetPollDetailFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/poll/detail", nil, true)

		return func(ctx context.Context, pollID int) (*GetPollDetailResponse, error) {
			payload := map[string]any{
				"poll_id": pollID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetPollDetail", err)
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

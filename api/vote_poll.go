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
	VotePollResponse struct {
		Options []model.PollOption `json:"options"`
	}
	VotePollFn = func(ctx context.Context, pollID string, optionID ...int) (*VotePollResponse, error)
)

func (a *api) VotePoll(ctx context.Context, pollID string, optionID ...int) (*VotePollResponse, error) {
	return a.e.VotePoll(ctx, pollID, optionID...)
}

var votePollFactory = apiFactory[*VotePollResponse, VotePollFn]()(
	func(a *api, sc session.Context, u factoryUtils[*VotePollResponse]) (VotePollFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/poll/vote", nil, true)

		return func(ctx context.Context, pollID string, optionID ...int) (*VotePollResponse, error) {
			payload := map[string]any{
				"poll_id":    pollID,
				"option_ids": optionID, // unvote = empty array
				"imei":       sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.VotePoll", err)
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

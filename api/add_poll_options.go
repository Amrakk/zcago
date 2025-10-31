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
	AddPollOptionsOption struct {
		Voted   bool   `json:"voted"`
		Content string `json:"content"`
	}
	AddPollOptionsRequest struct {
		PollId         int
		Options        []AddPollOptionsOption
		VotedOptionIds []int
	}
	AddPollOptionsResponse struct {
		Options []model.PollOption
	}
	AddPollOptionsFn = func(ctx context.Context, options AddPollOptionsRequest) (*AddPollOptionsResponse, error)
)

func (a *api) AddPollOptions(ctx context.Context, options AddPollOptionsRequest) (*AddPollOptionsResponse, error) {
	return a.e.AddPollOptions(ctx, options)
}

var addPollOptionsFactory = apiFactory[*AddPollOptionsResponse, AddPollOptionsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*AddPollOptionsResponse]) (AddPollOptionsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/poll/option/add", nil, true)

		return func(ctx context.Context, options AddPollOptionsRequest) (*AddPollOptionsResponse, error) {
			payload := map[string]any{
				"poll_id":          options.PollId,
				"new_options":      jsonx.Stringify(options.Options),
				"voted_option_ids": options.VotedOptionIds,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.AddPollOptions", err)
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

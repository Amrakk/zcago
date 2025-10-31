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
	CreatePollOptions struct {
		Question          string
		Options           []string
		ExpiredTime       int64
		AllowMultiChoices bool
		AllowAddNewOption bool
		HideVotePreview   bool
		IsAnonymous       bool
	}
	CreatePollResponse = model.PollDetail
	CreatePollFn       = func(ctx context.Context, groupID string, options CreatePollOptions) (*CreatePollResponse, error)
)

func (a *api) CreatePoll(ctx context.Context, groupID string, options CreatePollOptions) (*CreatePollResponse, error) {
	return a.e.CreatePoll(ctx, groupID, options)
}

var createPollFactory = apiFactory[*CreatePollResponse, CreatePollFn]()(
	func(a *api, sc session.Context, u factoryUtils[*CreatePollResponse]) (CreatePollFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/poll/create", nil, true)

		return func(ctx context.Context, groupID string, options CreatePollOptions) (*CreatePollResponse, error) {
			payload := map[string]any{
				"group_id":             groupID,
				"question":             options.Question,
				"options":              options.Options,
				"expired_time":         options.ExpiredTime,
				"pinAct":               false,
				"allow_multi_choices":  options.AllowMultiChoices,
				"allow_add_new_option": options.AllowAddNewOption,
				"is_hide_vote_preview": options.HideVotePreview,
				"is_anonymous":         options.IsAnonymous,
				"poll_type":            0,
				"src":                  1,
				"imei":                 sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.CreatePoll", err)
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

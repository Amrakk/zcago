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
	CreateAutoReplyRequest struct {
		Content   string               `json:"content"`
		IsEnable  bool                 `json:"isEnable"`
		StartTime int64                `json:"startTime"`
		EndTime   int64                `json:"endTime"`
		Scope     model.AutoReplyScope `json:"scope"`
		UIDs      []string             `json:"uids"`
	}
	CreateAutoReplyResponse struct {
		Item    model.AutoReplyItem `json:"item"`
		Version int                 `json:"version"`
	}

	CreateAutoReplyFn = func(ctx context.Context, message CreateAutoReplyRequest) (*CreateAutoReplyResponse, error)
)

func (a *api) CreateAutoReply(ctx context.Context, message CreateAutoReplyRequest) (*CreateAutoReplyResponse, error) {
	return a.e.CreateAutoReply(ctx, message)
}

var createAutoReplyFactory = apiFactory[*CreateAutoReplyResponse, CreateAutoReplyFn]()(
	func(a *api, sc session.Context, u factoryUtils[*CreateAutoReplyResponse]) (CreateAutoReplyFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("auto_reply"), "")
		serviceURL := u.MakeURL(base+"/api/autoreply/create", nil, true)

		return func(ctx context.Context, message CreateAutoReplyRequest) (*CreateAutoReplyResponse, error) {
			resultUids := make([]string, 0)
			if message.Scope == model.AutoReplyScopeSpecificFriends || message.Scope == model.AutoReplyScopeFriendsExcept {
				resultUids = message.UIDs
			}

			payload := map[string]any{
				"cliLang":    sc.Language(),
				"enable":     message.IsEnable,
				"content":    message.Content,
				"startTime":  message.StartTime,
				"endTime":    message.EndTime,
				"scope":      message.Scope,
				"recurrence": []string{"RRULE:FREQ=DAILY;"},
				"uids":       resultUids,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ChangeGroupName", err)
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

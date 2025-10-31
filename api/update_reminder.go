package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	UpdateReminderOptions struct {
		Title     string
		TopicId   string
		Emoji     string
		StartTime int64
		Repeat    model.ReminderRepeatMode
	}

	UpdateReminderUser  = model.ReminderUser
	UpdateReminderGroup struct {
		model.ReminderGroup
		ResponseMem *model.ResponseMembers `json:"responseMem,omitempty"`
	}

	UpdateReminderResponse = model.ReminderResponse[UpdateReminderUser, UpdateReminderGroup]
	UpdateReminderFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, options UpdateReminderOptions) (*UpdateReminderResponse, error)
)

func (a *api) UpdateReminder(ctx context.Context, threadID string, threadType model.ThreadType, options UpdateReminderOptions) (*UpdateReminderResponse, error) {
	return a.e.UpdateReminder(ctx, threadID, threadType, options)
}

var updateReminderFactory = apiFactory[*UpdateReminderResponse, UpdateReminderFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UpdateReminderResponse]) (UpdateReminderFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/board/oneone/update", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/board/topic/updatev2", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options UpdateReminderOptions) (*UpdateReminderResponse, error) {
			var payload map[string]any
			startTime := time.Now().UnixMilli()

			if options.StartTime != 0 {
				startTime = options.StartTime
			}

			if threadType == model.ThreadTypeUser {
				payload = map[string]any{
					"objectData": jsonx.Stringify(map[string]any{
						"toUid":      threadID,
						"type":       0,
						"color":      -16777216,
						"emoji":      options.Emoji,
						"startTime":  startTime,
						"duration":   -1,
						"params":     map[string]any{"title": options.Title},
						"needPin":    false,
						"reminderId": options.TopicId,
						"repeat":     options.Repeat,
					}),
					"imei": sc.IMEI(),
				}
			} else {
				payload = map[string]any{
					"grid":      threadID,
					"type":      0,
					"color":     -16777216,
					"emoji":     options.Emoji,
					"startTime": startTime,
					"duration":  -1,
					"params": jsonx.Stringify(map[string]any{
						"title": options.Title,
					}),
					"topicId": options.TopicId,
					"repeat":  options.Repeat,
					"imei":    sc.IMEI(),
					"pinAct":  2,
				}
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateReminder", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[threadType], &httpx.RequestOptions{
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

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
	CreateReminderOptions struct {
		Title     string
		Emoji     string
		StartTime int64
		Repeat    model.ReminderRepeatMode
	}

	CreateReminderUser  = model.ReminderUser
	CreateReminderGroup struct {
		model.ReminderGroup
		GroupID     *string                `json:"groupId,omitempty"`
		EventType   *int                   `json:"eventType,omitempty"`
		RepeatData  *[]any                 `json:"repeatData,omitempty"`
		ResponseMem *model.ResponseMembers `json:"responseMem,omitempty"`
	}

	CreateReminderResponse = model.ReminderResponse[CreateReminderUser, CreateReminderGroup]
	CreateReminderFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, options CreateReminderOptions) (*CreateReminderResponse, error)
)

func (a *api) CreateReminder(ctx context.Context, threadID string, threadType model.ThreadType, options CreateReminderOptions) (*CreateReminderResponse, error) {
	return a.e.CreateReminder(ctx, threadID, threadType, options)
}

var createReminderFactory = apiFactory[*CreateReminderResponse, CreateReminderFn]()(
	func(a *api, sc session.Context, u factoryUtils[*CreateReminderResponse]) (CreateReminderFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/board/oneone/create", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/board/topic/createv2", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options CreateReminderOptions) (*CreateReminderResponse, error) {
			var payload map[string]any
			emoji := "⏰"
			startTime := time.Now().UnixMilli()

			if len(options.Emoji) != 0 {
				emoji = options.Emoji
			}

			if options.StartTime != 0 {
				startTime = options.StartTime
			}

			if threadType == model.ThreadTypeUser {
				payload = map[string]any{
					"objectData": jsonx.Stringify(map[string]any{
						"toUid":      threadID,
						"type":       0,
						"color":      -16245706,
						"emoji":      emoji,
						"startTime":  startTime,
						"duration":   -1,
						"params":     map[string]any{"title": options.Title},
						"needPin":    false,
						"repeat":     options.Repeat,
						"creatorUid": sc.UID(), // Note: for some reason, you can put any valid UID here instead of your own and it still works, at least for mobile
						"src":        1,
					}),
					"imei": sc.IMEI(),
				}
			} else {
				payload = map[string]any{
					"grid":      threadID,
					"type":      0,
					"color":     -16245706,
					"emoji":     emoji,
					"startTime": startTime,
					"duration":  -1,
					"params": jsonx.Stringify(map[string]any{
						"title": options.Title,
					}),
					"repeat": options.Repeat,
					"src":    1,
					"imei":   sc.IMEI(),
					"pinAct": 0,
				}
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.CreateReminder", err)
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

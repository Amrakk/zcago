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
	RemoveReminderResponse = string
	RemoveReminderFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, reminderID string) (RemoveReminderResponse, error)
)

func (a *api) RemoveReminder(ctx context.Context, threadID string, threadType model.ThreadType, reminderID string) (RemoveReminderResponse, error) {
	return a.e.RemoveReminder(ctx, threadID, threadType, reminderID)
}

var removeReminderFactory = apiFactory[RemoveReminderResponse, RemoveReminderFn]()(
	func(a *api, sc session.Context, u factoryUtils[RemoveReminderResponse]) (RemoveReminderFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/board/oneone/remove", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/board/topic/remove", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, reminderID string) (RemoveReminderResponse, error) {
			var payload map[string]any

			if threadType == model.ThreadTypeUser {
				payload = map[string]any{
					"uid":        threadID,
					"reminderId": reminderID,
				}
			} else {
				payload = map[string]any{
					"grid":    threadID,
					"topicId": reminderID,
					"imei":    sc.IMEI(),
				}
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.RemoveReminder", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[threadType], &httpx.RequestOptions{
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

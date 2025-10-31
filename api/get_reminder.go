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
	GetReminderResponse = model.ReminderGroup
	GetReminderFn       = func(ctx context.Context, reminderID string) (*GetReminderResponse, error)
)

func (a *api) GetReminder(ctx context.Context, reminderID string) (*GetReminderResponse, error) {
	return a.e.GetReminder(ctx, reminderID)
}

var getReminderFactory = apiFactory[*GetReminderResponse, GetReminderFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetReminderResponse]) (GetReminderFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURL := u.MakeURL(base+"/api/board/topic/getReminder", nil, true)

		return func(ctx context.Context, reminderID string) (*GetReminderResponse, error) {
			payload := map[string]any{
				"eventId": reminderID,
				"imei":    sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetReminder", err)
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

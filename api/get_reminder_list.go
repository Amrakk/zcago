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
	ReminderListUser  = model.ReminderUser
	ReminderListGroup = model.ReminderGroup

	GetReminderListResponse = model.ReminderResponse[[]ReminderListUser, []ReminderListGroup]
	GetReminderListFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, options model.OffsetPaginationOptions) (*GetReminderListResponse, error)
)

func (a *api) GetReminderList(ctx context.Context, threadID string, threadType model.ThreadType, options model.OffsetPaginationOptions) (*GetReminderListResponse, error) {
	return a.e.GetReminderList(ctx, threadID, threadType, options)
}

var getReminderListFactory = apiFactory[*GetReminderListResponse, GetReminderListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetReminderListResponse]) (GetReminderListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/board/oneone/list", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/board/listReminder", nil, true),
		}
		return func(ctx context.Context, threadID string, threadType model.ThreadType, options model.OffsetPaginationOptions) (*GetReminderListResponse, error) {
			if options.Count <= 0 {
				options.Count = 20
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			objectData := map[string]any{
				"board_type": 1,
				"page":       options.Page,
				"count":      options.Count,
				"last_id":    0,
				"last_type":  0,
			}

			payload := map[string]any{}

			if threadType == model.ThreadTypeGroup {
				objectData["group_id"] = threadID
				payload["imei"] = sc.IMEI()
			} else {
				objectData["uid"] = threadID
			}

			payload["objectData"] = jsonx.Stringify(objectData)

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetReminderList", err)
			}

			url := u.MakeURL(serviceURLs[threadType], map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type UpdateSettingType string

const (
	// 0 = hide, 1 = show full (day/month/year), 2 = show day/month.
	UpdateViewBirthday UpdateSettingType = "view_birthday"
	// 0 = hide, 1 = show.
	UpdateShowOnlineStatus UpdateSettingType = "show_online_status"
	// 0 = hide, 1 = show.
	UpdateDisplaySeenStatus UpdateSettingType = "display_seen_status"
	// 1 = everyone, 2 = only friends.
	UpdateReceiveMessage UpdateSettingType = "receive_message"
	// 2 = only friends, 3 = everyone, 4 = friends and recent contacts.
	UpdateAcceptCall UpdateSettingType = "accept_stranger_call"
	// 0 = disable, 1 = enable.
	UpdateAddFriendViaPhone UpdateSettingType = "add_friend_via_phone"
	// 0 = disable, 1 = enable.
	UpdateAddFriendViaQR UpdateSettingType = "add_friend_via_qr"
	// 0 = disable, 1 = enable.
	UpdateAddFriendViaGroup UpdateSettingType = "add_friend_via_group"
	// 0 = disable, 1 = enable.
	UpdateAddFriendViaContact UpdateSettingType = "add_friend_via_contact"
	// 0 = disable, 1 = enable.
	UpdateDisplayOnRecommendFriend UpdateSettingType = "display_on_recommend_friend"
	// 0 = disable, 1 = enable.
	UpdateArchivedChat UpdateSettingType = "archivedChatStatus"
	// 0 = disable, 1 = enable.
	UpdateQuickMessage UpdateSettingType = "quickMessageStatus"
)

type (
	UpdateSettingResponse = string
	UpdateSettingFn       = func(ctx context.Context, sType UpdateSettingType, value int) (UpdateSettingResponse, error)
)

func (a *api) UpdateSetting(ctx context.Context, sType UpdateSettingType, value int) (UpdateSettingResponse, error) {
	return a.e.UpdateSetting(ctx, sType, value)
}

var updateSettingFactory = apiFactory[UpdateSettingResponse, UpdateSettingFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateSettingResponse]) (UpdateSettingFn, error) {
		serviceURL := u.MakeURL("https://wpa.chat.zalo.me/api/setting/update", nil, true)

		return func(ctx context.Context, sType UpdateSettingType, value int) (UpdateSettingResponse, error) {
			payload := map[string]any{
				string(sType): value,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateSetting", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

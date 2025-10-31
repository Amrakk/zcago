package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	UpdateGroupSettingOptions struct {
		BlockName        bool // Disallow group members to change the group name and avatar
		SignAdminMsg     bool // Highlight messages from owner/admins
		SetTopicOnly     bool // Don't pin messages, notes, and polls to the top of a conversation
		EnableMsgHistory bool // Allow new members to read most recent messages
		JoinApproval     bool // Require approval for new members to join the group
		LockCreatePost   bool // Disallow group members to create notes & reminders
		LockCreatePoll   bool // Disallow group members to create polls
		LockSendMsg      bool // Disallow group members to send messages
		LockViewMember   bool // Disallow group members to view full member list (community only)

		// BannFeature?: boolean; // not see in UI, not implemented
		// AddMemberOnly?: boolean; // not see in UI, not implemented
		// DirtyMedia?: boolean; // not see in UI, not implemented
		// BanDuration?: boolean | number; // not see in UI, not implemented
	}
	UpdateGroupSettingResponse = string
	UpdateGroupSettingFn       = func(ctx context.Context, groupID string, options UpdateGroupSettingOptions) (UpdateGroupSettingResponse, error)
)

func (a *api) UpdateGroupSetting(ctx context.Context, groupID string, options UpdateGroupSettingOptions) (UpdateGroupSettingResponse, error) {
	return a.e.UpdateGroupSetting(ctx, groupID, options)
}

var updateGroupSettingFactory = apiFactory[UpdateGroupSettingResponse, UpdateGroupSettingFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateGroupSettingResponse]) (UpdateGroupSettingFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/setting/update", nil, true)

		return func(ctx context.Context, groupID string, options UpdateGroupSettingOptions) (UpdateGroupSettingResponse, error) {
			payload := map[string]any{
				"blockName":    jsonx.B2I(options.BlockName),
				"signAdminMsg": jsonx.B2I(options.SignAdminMsg),

				// addMemberOnly: 0, // very tricky, any idea?

				"setTopicOnly":     jsonx.B2I(options.SetTopicOnly),
				"enableMsgHistory": jsonx.B2I(options.EnableMsgHistory),
				"joinAppr":         jsonx.B2I(options.JoinApproval),
				"lockCreatePost":   jsonx.B2I(options.LockCreatePost),
				"lockCreatePoll":   jsonx.B2I(options.LockCreatePoll),
				"lockSendMsg":      jsonx.B2I(options.LockSendMsg),
				"lockViewMember":   jsonx.B2I(options.LockViewMember),

				// default values for not implemented options
				"bannFeature":     0,
				"dirtyMedia":      0,
				"banDuration":     0,
				"blocked_members": []any{},

				"grid": groupID,
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateGroupSetting", err)
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

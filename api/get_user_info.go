package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type ProfileInfo = model.User

type GetUserInfoResponse struct {
	UnchangedProfiles map[string]any         `json:"unchanged_profiles"`
	PhonebookVersion  uint                   `json:"phonebook_version"`
	ChangedProfiles   map[string]ProfileInfo `json:"changed_profiles"`
}

type GetUserInfoFn = func(ctx context.Context, userId ...string) (*GetUserInfoResponse, error)

func (a *api) GetUserInfo(ctx context.Context, userID ...string) (*GetUserInfoResponse, error) {
	return a.e.GetUserInfo(ctx, userID...)
}

var getUserInfoFactory = apiFactory[*GetUserInfoResponse, GetUserInfoFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetUserInfoResponse]) (GetUserInfoFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/friend/getprofiles/v2", nil, true)

		return func(ctx context.Context, userID ...string) (*GetUserInfoResponse, error) {
			for i := range userID {
				if strings.IndexByte(userID[i], '_') < 0 {
					userID[i] += "_0"
				}
			}

			payload := map[string]any{
				"phonebook_version":   sc.ExtraVer().Phonebook,
				"friend_pversion_map": userID,
				"avatar_size":         120,
				"language":            sc.Language(),
				"show_online_status":  1,
				"imei":                sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetUserInfo", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{Method: http.MethodPost, Body: body})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, nil, true)
		}, nil
	},
)

package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/amrakk/zcago/errs"
	"github.com/amrakk/zcago/internal/httpx"
	"github.com/amrakk/zcago/internal/jsonx"
	"github.com/amrakk/zcago/model"
	"github.com/amrakk/zcago/session"
)

var ErrPhoneNumberEmpty = errs.NewZCA("phone number cannot be empty", "api.FindUser")

type (
	FindUserResponse map[string]struct {
		Avatar      string                 `json:"avatar"`
		Cover       string                 `json:"cover"`
		Status      string                 `json:"status"`
		Gender      model.Gender           `json:"gender"`
		DOB         int64                  `json:"dob"`
		Sdob        string                 `json:"sdob"`
		GlobalID    string                 `json:"globalId"`
		BizPkg      model.ZBusinessPackage `json:"bizPkg"`
		UID         string                 `json:"uid"`
		ZaloName    string                 `json:"zalo_name"`
		DisplayName string                 `json:"display_name"`
	}
	FindUserFn = func(ctx context.Context, avatarSize model.AvatarSize, phoneNumber ...string) (*FindUserResponse, error)
)

func (a *api) FindUser(ctx context.Context, phoneNumber ...string) (*FindUserResponse, error) {
	return a.e.FindUser(ctx, model.AvatarSizeLarge, phoneNumber...)
}

func (a *api) FindUserWithAvatarSize(ctx context.Context, avatarSize model.AvatarSize, phoneNumber ...string) (*FindUserResponse, error) {
	return a.e.FindUser(ctx, avatarSize, phoneNumber...)
}

var findUserFactory = apiFactory[*FindUserResponse, FindUserFn]()(
	func(a *api, sc session.Context, u factoryUtils[*FindUserResponse]) (FindUserFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/profile/multiget", nil, true)

		return func(ctx context.Context, avatarSize model.AvatarSize, phoneNumber ...string) (*FindUserResponse, error) {
			if !avatarSize.IsValid() {
				return nil, ErrInvalidAvatarSize
			}
			if len(phoneNumber) == 0 {
				return nil, ErrPhoneNumberEmpty
			}

			payload := map[string]any{
				"phones":      normalizePhoneNumbers(sc.Language(), phoneNumber),
				"avatar_size": avatarSize,
				"language":    sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.FindUser", err)
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

func normalizePhoneNumbers(language string, phoneNumbers []string) []string {
	if language != "vi" {
		return phoneNumbers
	}

	normalized := append([]string(nil), phoneNumbers...)
	for i, phone := range normalized {
		if strings.HasPrefix(phone, "0") {
			normalized[i] = "84" + phone[1:]
		}
	}
	return normalized
}

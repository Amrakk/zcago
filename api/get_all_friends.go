package api

import (
	"context"
	"net/http"

	"github.com/amrakk/zcago/errs"
	"github.com/amrakk/zcago/internal/httpx"
	"github.com/amrakk/zcago/internal/jsonx"
	"github.com/amrakk/zcago/model"
	"github.com/amrakk/zcago/session"
)

type (
	GetAllFriendsResponse []model.User
	GetAllFriendsFn       = func(ctx context.Context, options model.OffsetPaginationOptions, avatarSize model.AvatarSize) (*GetAllFriendsResponse, error)
)

func (a *api) GetAllFriends(ctx context.Context, options model.OffsetPaginationOptions) (*GetAllFriendsResponse, error) {
	return a.e.GetAllFriends(ctx, options, model.AvatarSizeSmall)
}

func (a *api) GetAllFriendsWithAvatarSize(ctx context.Context, options model.OffsetPaginationOptions, avatarSize model.AvatarSize) (*GetAllFriendsResponse, error) {
	return a.e.GetAllFriends(ctx, options, avatarSize)
}

var getAllFriendsFactory = apiFactory[*GetAllFriendsResponse, GetAllFriendsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAllFriendsResponse]) (GetAllFriendsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/friend/getfriends", nil, true)

		return func(ctx context.Context, options model.OffsetPaginationOptions, avatarSize model.AvatarSize) (*GetAllFriendsResponse, error) {
			if !avatarSize.IsValid() {
				return nil, ErrInvalidAvatarSize
			}
			if options.Count <= 0 {
				options.Count = 20000
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			payload := map[string]any{
				"page":        options.Page,
				"count":       options.Count,
				"incInvalid":  1,
				"avatar_size": avatarSize,
				"actiontime":  0,
				"imei":        sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetAllFriends", err)
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

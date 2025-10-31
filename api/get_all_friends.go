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
	GetAllFriendsResponse []model.User
	GetAllFriendsFn       = func(ctx context.Context, options model.OffsetPaginationOptions) (*GetAllFriendsResponse, error)
)

func (a *api) GetAllFriends(ctx context.Context, options model.OffsetPaginationOptions) (*GetAllFriendsResponse, error) {
	return a.e.GetAllFriends(ctx, options)
}

var getAllFriendsFactory = apiFactory[*GetAllFriendsResponse, GetAllFriendsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetAllFriendsResponse]) (GetAllFriendsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/friend/getfriends", nil, true)

		return func(ctx context.Context, options model.OffsetPaginationOptions) (*GetAllFriendsResponse, error) {
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
				"avatar_size": 120,
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

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
	AddGroupBlockedMemberResponse = string
	AddGroupBlockedMemberFn       = func(ctx context.Context, groupID string, memberID ...string) (AddGroupBlockedMemberResponse, error)
)

func (a *api) AddGroupBlockedMember(ctx context.Context, groupID string, memberID ...string) (AddGroupBlockedMemberResponse, error) {
	return a.e.AddGroupBlockedMember(ctx, groupID, memberID...)
}

var addGroupBlockedMemberFactory = apiFactory[AddGroupBlockedMemberResponse, AddGroupBlockedMemberFn]()(
	func(a *api, sc session.Context, u factoryUtils[AddGroupBlockedMemberResponse]) (AddGroupBlockedMemberFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/blockedmems/add", nil, true)

		return func(ctx context.Context, groupID string, memberID ...string) (AddGroupBlockedMemberResponse, error) {
			payload := map[string]any{
				"grid":    groupID,
				"members": memberID,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.AddGroupBlockedMember", err)
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

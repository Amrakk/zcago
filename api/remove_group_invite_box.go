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
	RemoveGroupInviteBoxResponse struct {
		DelInvitationIds []string     `json:"delInvitaionIds"`
		ErrMap           model.ErrMap `json:"errMap"`
	}
	RemoveGroupInviteBoxFn = func(ctx context.Context, blockFutureInvite bool, groupID ...string) (*RemoveGroupInviteBoxResponse, error)
)

func (a *api) RemoveGroupInviteBox(ctx context.Context, blockFutureInvite bool, groupID ...string) (*RemoveGroupInviteBoxResponse, error) {
	return a.e.RemoveGroupInviteBox(ctx, blockFutureInvite, groupID...)
}

var removeGroupInviteBoxFactory = apiFactory[*RemoveGroupInviteBoxResponse, RemoveGroupInviteBoxFn]()(
	func(a *api, sc session.Context, u factoryUtils[*RemoveGroupInviteBoxResponse]) (RemoveGroupInviteBoxFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/inv-box/mdel-inv", nil, true)

		return func(ctx context.Context, blockFutureInvite bool, groupID ...string) (*RemoveGroupInviteBoxResponse, error) {
			grids := make([]map[string]string, len(groupID))
			for i, id := range groupID {
				grids[i] = map[string]string{"grid": id}
			}

			payload := map[string]any{
				"invitations": jsonx.Stringify(grids),
				"block":       jsonx.B2I(blockFutureInvite),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.RemoveGroupInviteBox", err)
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

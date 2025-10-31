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
	GetFriendBoardListResponse struct {
		Data    []model.PinnedMessageDetail `json:"data"`
		Version int                         `json:"version"`
	}
	GetFriendBoardListFn = func(ctx context.Context, friendID string) (*GetFriendBoardListResponse, error)
)

func (a *api) GetFriendBoardList(ctx context.Context, friendID string) (*GetFriendBoardListResponse, error) {
	return a.e.GetFriendBoardList(ctx, friendID)
}

var getFriendBoardListFactory = apiFactory[*GetFriendBoardListResponse, GetFriendBoardListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetFriendBoardListResponse]) (GetFriendBoardListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend_board"), "")
		serviceURL := u.MakeURL(base+"/api/friendboard/list", nil, true)

		return func(ctx context.Context, friendID string) (*GetFriendBoardListResponse, error) {
			payload := map[string]any{
				"conversationId": friendID,
				"version":        0,
				"imei":           sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetFriendBoardList", err)
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

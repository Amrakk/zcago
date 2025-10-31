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
	MuteEntriesInfo struct {
		ID          string `json:"id"`
		Duration    int    `json:"duration"`
		StartTime   int64  `json:"startTime"`
		SystemTime  int64  `json:"systemTime"`
		CurrentTime int64  `json:"currentTime"`
		MuteMode    int    `json:"muteMode"`
	}
	GetMuteResponse struct {
		ChatEntries      []MuteEntriesInfo `json:"chatEntries"`
		GroupChatEntries []MuteEntriesInfo `json:"groupChatEntries"`
	}
	GetMuteFn = func(ctx context.Context) (*GetMuteResponse, error)
)

func (a *api) GetMute(ctx context.Context) (*GetMuteResponse, error) {
	return a.e.GetMute(ctx)
}

var getMuteFactory = apiFactory[*GetMuteResponse, GetMuteFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetMuteResponse]) (GetMuteFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/profile/getmute", nil, true)

		return func(ctx context.Context) (*GetMuteResponse, error) {
			payload := map[string]any{
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetMute", err)
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

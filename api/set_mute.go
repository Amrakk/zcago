package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type MuteAction int

const (
	MuteActionMute   MuteAction = 1
	MuteActionUnmute MuteAction = 3

	MuteDuration1Hour    int = 3600
	MuteDuration4Hour    int = 14400
	MuteDurationForever  int = -1
	MuteDurationUntil8AM int = 0
)

type (
	SetMuteOptions struct {
		Duration int // Mute duration in seconds.
		Action   MuteAction
	}
	SetMuteResponse = string
	SetMuteFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, options SetMuteOptions) (SetMuteResponse, error)
)

func (a *api) SetMute(ctx context.Context, threadID string, threadType model.ThreadType, options SetMuteOptions) (SetMuteResponse, error) {
	return a.e.SetMute(ctx, threadID, threadType, options)
}

var setMuteFactory = apiFactory[SetMuteResponse, SetMuteFn]()(
	func(a *api, sc session.Context, u factoryUtils[SetMuteResponse]) (SetMuteFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/profile/setmute", nil, true)

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options SetMuteOptions) (SetMuteResponse, error) {
			var muteDuration int

			if options.Action == MuteActionUnmute {
				muteDuration = -1
			} else if options.Duration == MuteDurationForever {
				muteDuration = -1
			} else if options.Duration == MuteDurationUntil8AM {
				now := time.Now()
				next8AM := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
				if now.After(next8AM) {
					next8AM = next8AM.Add(24 * time.Hour)
				}

				muteDuration = int(next8AM.Sub(now).Seconds())
			} else {
				muteDuration = options.Duration
			}

			payload := map[string]any{
				"toid":      threadID,
				"duration":  muteDuration,
				"action":    options.Action,
				"startTime": time.Now().Unix(),
				"muteType":  jsonx.B2I(threadType == model.ThreadTypeGroup) + 1,
				"imei":      sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SetMute", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

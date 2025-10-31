package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	UnreadMark struct {
		ID       int `json:"id"`
		CliMsgID int `json:"cliMsgId"`
		FromUID  int `json:"fromUid"`
		TS       int `json:"ts"`
	}
	GetUnreadMarkData struct {
		ConvsGroup []UnreadMark `json:"convsGroup"`
		ConvsUser  []UnreadMark `json:"convsUser"`
	}

	GetUnreadMarkResponse struct {
		Data   GetUnreadMarkData `json:"data"`
		Status int               `json:"status"`
	}
	GetUnreadMarkFn = func(ctx context.Context) (*GetUnreadMarkResponse, error)
)

func (a *api) GetUnreadMark(ctx context.Context) (*GetUnreadMarkResponse, error) {
	return a.e.GetUnreadMark(ctx)
}

var getUnreadMarkFactory = apiFactory[*GetUnreadMarkResponse, GetUnreadMarkFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetUnreadMarkResponse]) (GetUnreadMarkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/conv/getUnreadMark", nil, true)

		return func(ctx context.Context) (*GetUnreadMarkResponse, error) {
			payload := map[string]any{}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetUnreadMark", err)
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

func (r *GetUnreadMarkResponse) UnmarshalJSON(data []byte) error {
	type alias GetUnreadMarkResponse
	aux := &struct {
		Data string `json:"data"`
		*alias
	}{
		alias: (*alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Data == "" {
		r.Data = GetUnreadMarkData{}
		return nil
	}

	var mark GetUnreadMarkData
	if err := json.Unmarshal([]byte(aux.Data), &mark); err != nil {
		return err
	}
	r.Data = mark

	return nil
}

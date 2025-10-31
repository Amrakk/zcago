package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	RemoveUnreadMarkData struct {
		UpdateID int `json:"updateId"`
	}

	RemoveUnreadMarkResponse struct {
		Data   RemoveUnreadMarkData `json:"data"`
		Status int                  `json:"status"`
	}
	RemoveUnreadMarkFn = func(ctx context.Context, threadID string, threadType model.ThreadType) (*RemoveUnreadMarkResponse, error)
)

func (a *api) RemoveUnreadMark(ctx context.Context, threadID string, threadType model.ThreadType) (*RemoveUnreadMarkResponse, error) {
	return a.e.RemoveUnreadMark(ctx, threadID, threadType)
}

var removeUnreadMarkFactory = apiFactory[*RemoveUnreadMarkResponse, RemoveUnreadMarkFn]()(
	func(a *api, sc session.Context, u factoryUtils[*RemoveUnreadMarkResponse]) (RemoveUnreadMarkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/conv/removeUnreadMark", nil, true)

		return func(ctx context.Context, threadID string, threadType model.ThreadType) (*RemoveUnreadMarkResponse, error) {
			now := time.Now().UnixMilli()

			activeConvsKey, inactiveConvsKey := "convsUser", "convsGroup"
			activeDataKey, inactiveDataKey := "convsUserData", "convsGroupData"
			if threadType == model.ThreadTypeGroup {
				activeConvsKey, inactiveConvsKey = inactiveConvsKey, activeConvsKey
				activeDataKey, inactiveDataKey = inactiveDataKey, activeDataKey
			}

			inner := map[string]any{
				activeConvsKey:   []string{threadID},
				inactiveConvsKey: []string{},
				activeDataKey: []map[string]any{
					{"id": threadID, "ts": now},
				},
				inactiveDataKey: []any{},
			}

			payload := map[string]any{
				"param": jsonx.Stringify(inner),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.RemoveUnreadMark", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{
				Method: http.MethodPost,
				Body:   body,
			})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

func (r *RemoveUnreadMarkResponse) UnmarshalJSON(data []byte) error {
	type alias RemoveUnreadMarkResponse
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
		r.Data = RemoveUnreadMarkData{}
		return nil
	}

	var mark RemoveUnreadMarkData
	if err := json.Unmarshal([]byte(aux.Data), &mark); err != nil {
		return err
	}
	r.Data = mark

	return nil
}

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	AddUnreadMarkData struct {
		UpdateID int `json:"updateId"`
	}
	AddUnreadMarkResponse struct {
		Data   AddUnreadMarkData `json:"data"`
		Status int               `json:"status"`
	}
	AddUnreadMarkFn = func(ctx context.Context, threadID string, threadType model.ThreadType) (*AddUnreadMarkResponse, error)
)

func (a *api) AddUnreadMark(ctx context.Context, threadID string, threadType model.ThreadType) (*AddUnreadMarkResponse, error) {
	return a.e.AddUnreadMark(ctx, threadID, threadType)
}

var addUnreadMarkFactory = apiFactory[*AddUnreadMarkResponse, AddUnreadMarkFn]()(
	func(a *api, sc session.Context, u factoryUtils[*AddUnreadMarkResponse]) (AddUnreadMarkFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("conversation"), "")
		serviceURL := u.MakeURL(base+"/api/conv/addUnreadMark", nil, true)

		return func(ctx context.Context, threadID string, threadType model.ThreadType) (*AddUnreadMarkResponse, error) {
			now := time.Now().UnixMilli()
			cliMsgID := strconv.FormatInt(now, 10)

			activeKey := "convsUser"
			inactiveKey := "convsGroup"
			if threadType == model.ThreadTypeGroup {
				activeKey, inactiveKey = inactiveKey, activeKey
			}

			param := map[string]any{
				activeKey: []map[string]any{
					{
						"id":       threadID,
						"cliMsgId": cliMsgID,
						"fromUid":  "0",
						"ts":       now,
					},
				},
				inactiveKey: []any{},
				"imei":      sc.IMEI(),
			}

			payload := map[string]any{
				"param": jsonx.Stringify(param),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ForwardMessage", err)
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

func (r *AddUnreadMarkResponse) UnmarshalJSON(data []byte) error {
	type alias AddUnreadMarkResponse
	aux := &struct {
		Data string `json:"data"`
		*alias
	}{
		alias: (*alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var dataObj AddUnreadMarkData
	if err := json.Unmarshal([]byte(aux.Data), &dataObj); err != nil {
		return err
	}
	r.Data = dataObj
	return nil
}

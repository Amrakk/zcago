package api

import (
	"context"
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
	SendCardOptions struct {
		UserID      string
		PhoneNumber string
		TTL         int
	}
	SendCardResponse struct {
		MsgID int `json:"msgId"`
	}
	SendCardFn = func(ctx context.Context, threadID string, threadType model.ThreadType, options SendCardOptions) (*SendCardResponse, error)
)

func (a *api) SendCard(ctx context.Context, threadID string, threadType model.ThreadType, options SendCardOptions) (*SendCardResponse, error) {
	return a.e.SendCard(ctx, threadID, threadType, options)
}

var sendCardFactory = apiFactory[*SendCardResponse, SendCardFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendCardResponse]) (SendCardFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/forward", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/forward", nil, true),
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, options SendCardOptions) (*SendCardResponse, error) {
			data, err := a.GetQR(ctx, options.UserID)
			if err != nil {
				return nil, err
			}

			msgInfo := map[string]any{
				"contactUid": options.UserID,
				"qrCodeUrl":  (*data)[options.UserID],
			}

			if options.PhoneNumber != "" {
				msgInfo["phone"] = options.PhoneNumber
			}

			payload := map[string]any{
				"ttl":      options.TTL,
				"msgType":  6,
				"clientId": strconv.FormatInt(time.Now().UnixMilli(), 10),
				"msgInfo":  jsonx.Stringify(msgInfo),
			}

			if threadType == model.ThreadTypeUser {
				payload["toId"] = threadID
				payload["imei"] = sc.IMEI()
			} else {
				payload["grid"] = threadID
				payload["visibility"] = 0
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendCard", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[threadType], &httpx.RequestOptions{
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

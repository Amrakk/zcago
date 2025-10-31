package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	SendBankCardData struct {
		BinBank     model.BinBankCard
		NumAccBank  string
		NameAccBank string
	}
	SendBankCardResponse = string
	SendBankCardFn       = func(ctx context.Context, threadID string, threadType model.ThreadType, data SendBankCardData) (SendBankCardResponse, error)
)

func (a *api) SendBankCard(ctx context.Context, threadID string, threadType model.ThreadType, data SendBankCardData) (SendBankCardResponse, error) {
	return a.e.SendBankCard(ctx, threadID, threadType, data)
}

var sendBankCardFactory = apiFactory[SendBankCardResponse, SendBankCardFn]()(
	func(a *api, sc session.Context, u factoryUtils[SendBankCardResponse]) (SendBankCardFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("zimsg"), "")
		serviceURL := u.MakeURL(base+"/api/transfer/card", nil, true)

		return func(ctx context.Context, threadID string, threadType model.ThreadType, data SendBankCardData) (SendBankCardResponse, error) {
			now := time.Now().UnixMilli()
			nameAccBank := data.NameAccBank

			if len(nameAccBank) == 0 {
				nameAccBank = "---"
			}

			payload := map[string]any{
				"binBank":     data.BinBank,
				"numAccBank":  data.NumAccBank,
				"nameAccBank": strings.ToUpper(nameAccBank),
				"cliMsgId":    strconv.FormatInt(now, 10),
				"tsMsg":       now,
				"destUid":     threadID,
				"destType":    jsonx.B2I(threadType == model.ThreadTypeGroup),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.SendBankCard", err)
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

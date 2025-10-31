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

var ErrInvalidReaction = errs.NewZCA("invalid reaction data", "api.AddReaction")

type (
	AddReactionData struct {
		MsgID    string
		CliMsgID string
	}
	AddReactionDestination struct {
		ThreadID string
		Type     model.ThreadType
		Data     AddReactionData
	}

	AddReactionResponse struct {
		MsgIDs []int `json:"msgIds"`
	}
	AddReactionFn = func(ctx context.Context, dest AddReactionDestination, reaction model.ReactionData) (*AddReactionResponse, error)
)

func (a *api) AddReaction(ctx context.Context, dest AddReactionDestination, reaction model.ReactionData) (*AddReactionResponse, error) {
	return a.e.AddReaction(ctx, dest, reaction)
}

var addReactionFactory = apiFactory[*AddReactionResponse, AddReactionFn]()(
	func(a *api, sc session.Context, u factoryUtils[*AddReactionResponse]) (AddReactionFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("reaction"), "")
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/reaction", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/reaction", nil, true),
		}

		return func(ctx context.Context, dest AddReactionDestination, reaction model.ReactionData) (*AddReactionResponse, error) {
			if !reaction.IsValid() {
				return nil, ErrInvalidReaction
			}

			gMsgID, _ := strconv.ParseInt(dest.Data.MsgID, 10, 64)
			cMsgID, _ := strconv.ParseInt(dest.Data.CliMsgID, 10, 64)

			msg := map[string]any{
				"rMsg": []map[string]any{{
					"gMsgID":  gMsgID,
					"cMsgID":  cMsgID,
					"msgType": 1,
				}},
				"rIcon":  reaction.RIcon,
				"rType":  reaction.RType,
				"source": reaction.Source,
			}

			reactList := []map[string]any{{
				"message":  jsonx.Stringify(msg),
				"clientId": time.Now().UnixMilli(),
			}}

			payload := map[string]any{
				"react_list": reactList,
			}

			if dest.Type == model.ThreadTypeUser {
				payload["toid"] = dest.ThreadID
			} else {
				payload["grid"] = dest.ThreadID
				payload["imei"] = sc.IMEI()
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.AddReaction", err)
			}

			body := httpx.BuildFormBody(map[string]string{"params": enc})
			resp, err := u.Request(ctx, serviceURLs[dest.Type], &httpx.RequestOptions{
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

func (r *AddReactionResponse) UnmarshalJSON(data []byte) error {
	type alias AddReactionResponse
	aux := &struct {
		MsgIDs string `json:"msgIds"`
		*alias
	}{
		alias: (*alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.MsgIDs == "" {
		r.MsgIDs = []int{}
		return nil
	}

	var ids []int
	if err := json.Unmarshal([]byte(aux.MsgIDs), &ids); err != nil {
		return err
	}
	r.MsgIDs = ids

	return nil
}

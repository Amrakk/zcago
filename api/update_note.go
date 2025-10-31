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
	UpdateNoteOptions struct {
		Title   string
		TopicId string
		PinAct  bool
	}
	UpdateNoteResponse = model.NoteDetail
	UpdateNoteFn       = func(ctx context.Context, groupID string, options UpdateNoteOptions) (*UpdateNoteResponse, error)
)

func (a *api) UpdateNote(ctx context.Context, groupID string, options UpdateNoteOptions) (*UpdateNoteResponse, error) {
	return a.e.UpdateNote(ctx, groupID, options)
}

var updateNoteFactory = apiFactory[*UpdateNoteResponse, UpdateNoteFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UpdateNoteResponse]) (UpdateNoteFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURL := u.MakeURL(base+"/api/board/topic/updatev2", nil, true)

		return func(ctx context.Context, groupID string, options UpdateNoteOptions) (*UpdateNoteResponse, error) {
			payload := map[string]any{
				"grid":      groupID,
				"type":      0,
				"color":     -16777216,
				"emoji":     "",
				"startTime": -1,
				"duration":  -1,
				"params": jsonx.Stringify(map[string]any{
					"title": options.Title,
				}),
				"topicId": options.TopicId,
				"repeat":  0,
				"imei":    sc.IMEI(),
				"pinAct":  jsonx.B2I(options.PinAct) + 1,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateNote", err)
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

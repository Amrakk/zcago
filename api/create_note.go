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
	CreateNoteOptions struct {
		Title  string `json:"title"`
		PinAct bool   `json:"pinAct"`
	}
	CreateNoteResponse = model.NoteDetail
	CreateNoteFn       = func(ctx context.Context, groupID string, options CreateNoteOptions) (*CreateNoteResponse, error)
)

func (a *api) CreateNote(ctx context.Context, groupID string, options CreateNoteOptions) (*CreateNoteResponse, error) {
	return a.e.CreateNote(ctx, groupID, options)
}

var createNoteFactory = apiFactory[*CreateNoteResponse, CreateNoteFn]()(
	func(a *api, sc session.Context, u factoryUtils[*CreateNoteResponse]) (CreateNoteFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURL := u.MakeURL(base+"/api/board/topic/createv2", nil, true)

		return func(ctx context.Context, groupID string, options CreateNoteOptions) (*CreateNoteResponse, error) {
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
				"repeat": 0,
				"src":    1,
				"imei":   sc.IMEI(),
				"pinAct": jsonx.B2I(options.PinAct),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.CreateNote", err)
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

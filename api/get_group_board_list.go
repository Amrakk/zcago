package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	BoardItem struct {
		BoardType model.BoardType `json:"boardType"`
		Data      model.BoardData `json:"data"` // model.PollDetail | model.NoteDetail | model.PinnedMessageDetail
	}

	GetGroupBoardListResponse struct {
		Items []BoardItem `json:"items"`
		Count int         `json:"count"`
	}
	GetGroupBoardListFn = func(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupBoardListResponse, error)
)

func (a *api) GetGroupBoardList(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupBoardListResponse, error) {
	return a.e.GetGroupBoardList(ctx, groupID, options)
}

var getGroupBoardListFactory = apiFactory[*GetGroupBoardListResponse, GetGroupBoardListFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetGroupBoardListResponse]) (GetGroupBoardListFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group_board"), "")
		serviceURL := u.MakeURL(base+"/api/board/list", nil, true)

		return func(ctx context.Context, groupID string, options model.OffsetPaginationOptions) (*GetGroupBoardListResponse, error) {
			if options.Count <= 0 {
				options.Count = 20
			}
			if options.Page <= 0 {
				options.Page = 1
			}

			payload := map[string]any{
				"group_id":   groupID,
				"board_type": 0,
				"page":       options.Page,
				"count":      options.Count,
				"last_id":    0,
				"last_type":  0,
				"imei":       sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetGroupBoardList", err)
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

func (bi *BoardItem) UnmarshalJSON(data []byte) error {
	type alias struct {
		BoardType model.BoardType `json:"boardType"`
		Data      json.RawMessage `json:"data"`
	}

	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	bi.BoardType = aux.BoardType

	switch aux.BoardType {
	case model.BoardTypePoll:
		var poll model.PollDetail
		if err := json.Unmarshal(aux.Data, &poll); err != nil {
			return err
		}
		bi.Data = &poll
	case model.BoardTypeNote:
		var note model.NoteDetail
		if err := json.Unmarshal(aux.Data, &note); err != nil {
			return err
		}
		bi.Data = &note
	case model.BoardTypePinnedMessage:
		var pinned model.PinnedMessageDetail
		if err := json.Unmarshal(aux.Data, &pinned); err != nil {
			return err
		}
		bi.Data = &pinned
	default:
		return fmt.Errorf("unknown board type: %d", aux.BoardType)
	}

	return nil
}

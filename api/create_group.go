package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	CreateGroupOptions struct {
		Name    string
		Members []string
	}
	CreateGroupResponse struct {
		GroupType      model.GroupType `json:"groupType"`
		SuccessMembers []string        `json:"sucessMembers"`
		GroupID        string          `json:"groupId"`
		ErrorMembers   []string        `json:"errorMembers"`
		ErrorData      map[string]any  `json:"error_data"`
	}
	CreateGroupFn = func(ctx context.Context, options CreateGroupOptions) (*CreateGroupResponse, error)
)

func (a *api) CreateGroup(ctx context.Context, options CreateGroupOptions) (*CreateGroupResponse, error) {
	return a.e.CreateGroup(ctx, options)
}

var createGroupFactory = apiFactory[*CreateGroupResponse, CreateGroupFn]()(
	func(a *api, sc session.Context, u factoryUtils[*CreateGroupResponse]) (CreateGroupFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/create/v2", nil, true)

		return func(ctx context.Context, options CreateGroupOptions) (*CreateGroupResponse, error) {
			if len(options.Members) == 0 {
				return nil, errs.NewZCA("members cannot be empty", "api.CreateGroup")
			}

			now := time.Now().UnixMilli()
			gname := fmt.Sprint(now)
			nameChanged := 0
			memberTypes := make([]int, len(options.Members))
			for i := range options.Members {
				memberTypes[i] = -1
			}

			if len(options.Name) > 0 {
				nameChanged = 1
				gname = options.Name
			}

			payload := map[string]any{
				"clientId":     now,
				"gname":        gname,
				"gdesc":        nil,
				"members":      options.Members,
				"membersTypes": memberTypes,
				"nameChanged":  nameChanged,
				"createLink":   1,
				"imei":         sc.IMEI(),
				"clientLang":   sc.Language(),
				"zsource":      601,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ChangeGroupOwner", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{Method: http.MethodPost})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type ReviewPendingMemberRequestStatus int

const (
	ReviewStatusSuccess                ReviewPendingMemberRequestStatus = 0
	ReviewStatusNotInPendingList       ReviewPendingMemberRequestStatus = 170
	ReviewStatusAlreadyInGroup         ReviewPendingMemberRequestStatus = 178
	ReviewStatusInsufficientPermission ReviewPendingMemberRequestStatus = 166
)

type (
	ReviewPendingMemberData struct {
		Members   []string
		IsApprove bool
	}
	ReviewPendingMemberRequestResponse map[string]ReviewPendingMemberRequestStatus
	ReviewPendingMemberRequestFn       = func(ctx context.Context, groupID string, data ReviewPendingMemberData) (*ReviewPendingMemberRequestResponse, error)
)

func (a *api) ReviewPendingMemberRequest(ctx context.Context, groupID string, data ReviewPendingMemberData) (*ReviewPendingMemberRequestResponse, error) {
	return a.e.ReviewPendingMemberRequest(ctx, groupID, data)
}

var reviewPendingMemberRequestFactory = apiFactory[*ReviewPendingMemberRequestResponse, ReviewPendingMemberRequestFn]()(
	func(a *api, sc session.Context, u factoryUtils[*ReviewPendingMemberRequestResponse]) (ReviewPendingMemberRequestFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("group"), "")
		serviceURL := u.MakeURL(base+"/api/group/pending-mems/review", nil, true)

		return func(ctx context.Context, groupID string, data ReviewPendingMemberData) (*ReviewPendingMemberRequestResponse, error) {
			payload := map[string]any{
				"grid":      groupID,
				"members":   data.Members,
				"isApprove": jsonx.B2I(data.IsApprove),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.ReviewPendingMemberRequest", err)
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

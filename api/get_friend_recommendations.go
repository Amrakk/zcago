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
	CollapseConfig struct {
		CollapseID    int `json:"collapseId"`
		CollapseXItem int `json:"collapseXItem"`
		CollapseYItem int `json:"collapseYItem"`
	}
	RecommendationMeta struct {
		SuggestWay int     `json:"suggestWay"`
		Source     int     `json:"source"`
		Message    string  `json:"message"`
		CustomText *string `json:"customText"` // nullable in JSON
	}
	FriendRecommendation struct {
		UserID          string                 `json:"userId"`
		ZaloName        string                 `json:"zaloName"`
		DisplayName     string                 `json:"displayName"`
		Avatar          string                 `json:"avatar"`
		PhoneNumber     string                 `json:"phoneNumber"`
		Status          string                 `json:"status"`
		Gender          model.Gender           `json:"gender"`
		Dob             int64                  `json:"dob"`
		Type            int                    `json:"type"`
		RecommType      int                    `json:"recommType"`
		RecommSrc       int                    `json:"recommSrc"`
		RecommTime      int64                  `json:"recommTime"`
		RecommInfo      RecommendationMeta     `json:"recommInfo"`
		BizPkg          model.ZBusinessPackage `json:"bizPkg"`
		IsSeenFriendReq bool                   `json:"isSeenFriendReq"`
	}
	RecommendationItem struct {
		RecommItemType int                  `json:"recommItemType"`
		DataInfo       FriendRecommendation `json:"dataInfo"`
	}

	GetFriendRecommendationsResponse struct {
		ExpiredDuration       int64                `json:"expiredDuration"`
		CollapseMsgListConfig CollapseConfig       `json:"collapseMsgListConfig"`
		RecommItems           []RecommendationItem `json:"recommItems"`
	}
	GetFriendRecommendationsFn = func(ctx context.Context) (*GetFriendRecommendationsResponse, error)
)

func (a *api) GetFriendRecommendations(ctx context.Context) (*GetFriendRecommendationsResponse, error) {
	return a.e.GetFriendRecommendations(ctx)
}

var getFriendRecommendationsFactory = apiFactory[*GetFriendRecommendationsResponse, GetFriendRecommendationsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetFriendRecommendationsResponse]) (GetFriendRecommendationsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("friend"), "")
		serviceURL := u.MakeURL(base+"/api/friend/recommendsv2/list", nil, true)

		return func(ctx context.Context) (*GetFriendRecommendationsResponse, error) {
			payload := map[string]any{
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetFriendRecommendations", err)
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

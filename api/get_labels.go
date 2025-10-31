package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	GetLabelsResponse struct {
		LabelData      []model.LabelData `json:"labelData"`
		Version        int               `json:"version"`
		LastUpdateTime int64             `json:"lastUpdateTime"`
	}
	GetLabelsFn = func(ctx context.Context) (*GetLabelsResponse, error)
)

func (a *api) GetLabels(ctx context.Context) (*GetLabelsResponse, error) {
	return a.e.GetLabels(ctx)
}

var getLabelsFactory = apiFactory[*GetLabelsResponse, GetLabelsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*GetLabelsResponse]) (GetLabelsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("label"), "")
		serviceURL := u.MakeURL(base+"/api/convlabel/get", nil, true)

		return func(ctx context.Context) (*GetLabelsResponse, error) {
			payload := map[string]any{
				"imei": sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.GetLabels", err)
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

func (r *GetLabelsResponse) UnmarshalJSON(data []byte) error {
	type alias GetLabelsResponse
	aux := &struct {
		LabelData string `json:"labelData"`
		*alias
	}{
		alias: (*alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var labels []model.LabelData
	if err := json.Unmarshal([]byte(aux.LabelData), &labels); err != nil {
		return err
	}

	r.LabelData = labels
	return nil
}

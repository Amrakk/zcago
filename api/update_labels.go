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
	UpdateLabelsData struct {
		LabelData []model.LabelData
		Version   int
	}
	UpdateLabelsResponse struct {
		LabelData      []model.LabelData
		Version        int
		LastUpdateTime int64
	}
	UpdateLabelsFn = func(ctx context.Context, labels UpdateLabelsData) (*UpdateLabelsResponse, error)
)

func (a *api) UpdateLabels(ctx context.Context, labels UpdateLabelsData) (*UpdateLabelsResponse, error) {
	return a.e.UpdateLabels(ctx, labels)
}

var updateLabelsFactory = apiFactory[*UpdateLabelsResponse, UpdateLabelsFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UpdateLabelsResponse]) (UpdateLabelsFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("label"), "")
		serviceURL := u.MakeURL(base+"/api/convlabel/update", nil, true)

		return func(ctx context.Context, labels UpdateLabelsData) (*UpdateLabelsResponse, error) {
			payload := map[string]any{
				"labelData": jsonx.Stringify(labels.LabelData),
				"version":   labels.Version,
				"imei":      sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UpdateLabels", err)
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

func (r *UpdateLabelsResponse) UnmarshalJSON(data []byte) error {
	type alias UpdateLabelsResponse
	aux := &struct {
		LabelData string `json:"labelData"`
		*alias
	}{
		alias: (*alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.LabelData == "" {
		r.LabelData = []model.LabelData{}
		return nil
	}

	var labels []model.LabelData
	if err := json.Unmarshal([]byte(aux.LabelData), &labels); err != nil {
		return err
	}
	r.LabelData = labels

	return nil
}

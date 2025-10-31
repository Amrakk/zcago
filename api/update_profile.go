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
	UpdateProfileProfile struct {
		Name   string       `json:"name"`
		DOB    string       `json:"dob"` // Date of Birth in format "YYYY-MM-DD"
		Gender model.Gender `json:"gender"`
	}
	UpdateProfileBiz struct {
		Category    model.BusinessCategory `json:"cate,omitempty"`
		Description string                 `json:"desc,omitempty"`
		Address     string                 `json:"addr,omitempty"`
		Website     string                 `json:"website,omitempty"`
		Email       string                 `json:"email,omitempty"`
	}
	UpdateProfileData struct {
		Profile UpdateProfileProfile
		Biz     UpdateProfileBiz
	}
	UpdateProfileResponse = string
	UpdateProfileFn       = func(ctx context.Context, data UpdateProfileData) (UpdateProfileResponse, error)
)

func (a *api) UpdateProfile(ctx context.Context, data UpdateProfileData) (UpdateProfileResponse, error) {
	return a.e.UpdateProfile(ctx, data)
}

var updateProfileFactory = apiFactory[UpdateProfileResponse, UpdateProfileFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateProfileResponse]) (UpdateProfileFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("profile"), "")
		serviceURL := u.MakeURL(base+"/api/social/profile/update", nil, true)

		return func(ctx context.Context, data UpdateProfileData) (UpdateProfileResponse, error) {
			payload := map[string]any{
				"profile":  jsonx.Stringify(data.Profile),
				"biz":      jsonx.Stringify(data.Biz),
				"language": sc.Language(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateProfile", err)
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

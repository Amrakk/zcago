package api

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
)

type (
	UploadThumbnailResponse struct {
		HDURL        string `json:"hdUrl"`
		URL          string `json:"url"`
		ClientFileID int    `json:"clientFileId"`
		FileID       int    `json:"fileId"`
	}
	UploadThumbnailFn = func(ctx context.Context, source model.AttachmentSource) (*UploadThumbnailResponse, error)
)

func (a *api) UploadThumbnail(ctx context.Context, source model.AttachmentSource) (*UploadThumbnailResponse, error) {
	return a.e.UploadThumbnail(ctx, source)
}

var uploadThumbnailFactory = apiFactory[*UploadThumbnailResponse, UploadThumbnailFn]()(
	func(a *api, sc session.Context, u factoryUtils[*UploadThumbnailResponse]) (UploadThumbnailFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURL := u.MakeURL(base+"/api/message/upthumb", nil, true)

		return func(ctx context.Context, source model.AttachmentSource) (*UploadThumbnailResponse, error) {
			var (
				reader io.Reader
				closer io.Closer
			)

			defer func() {
				if closer != nil {
					_ = closer.Close()
				}
			}()

			if f := source.String(); f != "" {
				r, err := os.Open(f)
				if err != nil {
					return nil, errs.WrapZCA("failed to read file", "api.UploadThumbnail", err)
				}

				reader = r
				closer = r

			} else if f := source.Object(); f != nil {
				reader = f.Data
			}

			forms, err := httpx.BuildFormData("fileContent", reader, httpx.WithContentType("image/png"))
			if err != nil || len(forms) == 0 || forms[0] == nil {
				return nil, errs.WrapZCA("failed to build form data", "api.UploadPhoto", err)
			}
			formData := forms[0]

			payload := map[string]any{
				"clientId": time.Now().UnixMilli(),
				"imei":     sc.IMEI(),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.UploadThumbnail", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{
				Method:  http.MethodPost,
				Headers: formData.Header,
				Body:    formData.Body,
			})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

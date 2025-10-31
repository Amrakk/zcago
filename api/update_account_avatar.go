package api

import (
	"context"
	"fmt"
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
	UpdateAccountAvatarResponse = string
	UpdateAccountAvatarFn       = func(ctx context.Context, source model.AttachmentSource) (UpdateAccountAvatarResponse, error)
)

func (a *api) UpdateAccountAvatar(ctx context.Context, source model.AttachmentSource) (UpdateAccountAvatarResponse, error) {
	return a.e.UpdateAccountAvatar(ctx, source)
}

var updateAccountAvatarFactory = apiFactory[UpdateAccountAvatarResponse, UpdateAccountAvatarFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateAccountAvatarResponse]) (UpdateAccountAvatarFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURL := u.MakeURL(base+"/api/profile/upavatar", nil, true)

		return func(ctx context.Context, source model.AttachmentSource) (UpdateAccountAvatarResponse, error) {
			var (
				reader       io.Reader
				closer       io.Closer
				fileName     string
				fileMetadata model.AttachmentMetadata
			)

			defer func() {
				if closer != nil {
					_ = closer.Close()
				}
			}()

			if f := source.String(); f != "" {
				r, err := os.Open(f)
				if err != nil {
					return "", errs.WrapZCA("failed to read file", "api.UpdateAccountAvatar", err)
				}

				closer = r
				reader = r

				fileMetadata, fileName, err = sc.GetImageMetadata(f)
				if err != nil {
					return "", err
				}
			} else if f := source.Object(); f != nil {
				reader, fileName, fileMetadata = f.Data, f.Filename, f.Metadata
			}

			forms, err := httpx.BuildFormData(
				"fileContent", reader,
				httpx.WithContentType("image/jpeg"),
				httpx.WithFileName(fileName),
			)
			if err != nil || len(forms) == 0 || forms[0] == nil {
				return "", errs.WrapZCA("failed to build form data", "api.UpdateAccountAvatar", err)
			}
			formData := forms[0]

			metaWidth := jsonx.Or(fileMetadata.Width, 1080)
			metaHeight := jsonx.Or(fileMetadata.Height, 1080)
			metaSize := jsonx.Or(fileMetadata.Size, 0)

			common := map[string]int{
				"width":  metaWidth,
				"height": metaHeight,
			}

			metadata := jsonx.Stringify(map[string]any{
				"origin": common,
				"processed": map[string]any{
					"width":  metaWidth,
					"height": metaHeight,
					"size":   metaSize,
				},
			})

			now := time.Now()
			clientID := fmt.Sprintf("%s %s", sc.UID(), now.Format("15:04 02/01/2006"))

			payload := map[string]any{
				"avatarSize": 120,
				"clientId":   clientID,
				"language":   sc.Language(),
				"metaData":   metadata,
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateAccountAvatar", err)
			}

			url := u.MakeURL(serviceURL, map[string]any{"params": enc}, true)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{
				Method:  http.MethodPost,
				Headers: formData.Header,
				Body:    formData.Body,
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, true)
		}, nil
	},
)

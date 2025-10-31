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
	UpdateGroupAvatarResponse = string
	UpdateGroupAvatarFn       = func(ctx context.Context, groupID string, source model.AttachmentSource) (UpdateGroupAvatarResponse, error)
)

func (a *api) UpdateGroupAvatar(ctx context.Context, groupID string, source model.AttachmentSource) (UpdateGroupAvatarResponse, error) {
	return a.e.UpdateGroupAvatar(ctx, groupID, source)
}

var updateGroupAvatarFactory = apiFactory[UpdateGroupAvatarResponse, UpdateGroupAvatarFn]()(
	func(a *api, sc session.Context, u factoryUtils[UpdateGroupAvatarResponse]) (UpdateGroupAvatarFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		serviceURL := u.MakeURL(base+"/api/group/upavatar", nil, true)

		return func(ctx context.Context, groupID string, source model.AttachmentSource) (UpdateGroupAvatarResponse, error) {
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
					return "", errs.WrapZCA("failed to read file", "api.UpdateGroupAvatar", err)
				}

				reader = r
				closer = r

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
				return "", errs.WrapZCA("failed to build form data", "api.UpdateGroupAvatar", err)
			}
			formData := forms[0]

			now := time.Now()
			clientID := fmt.Sprintf("g%s%s", groupID, now.Format("15:04 02/01/2006"))

			payload := map[string]any{
				"grid":         groupID,
				"avatarSize":   120,
				"clientId":     clientID,
				"imei":         sc.IMEI(),
				"originWidth":  jsonx.Or(fileMetadata.Width, 1080),
				"originHeight": jsonx.Or(fileMetadata.Height, 1080),
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return "", errs.WrapZCA("failed to encrypt params", "api.UpdateGroupAvatar", err)
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

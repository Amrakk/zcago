package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/amrakk/zcago/errs"
	"github.com/amrakk/zcago/internal/httpx"
	"github.com/amrakk/zcago/internal/jsonx"
	"github.com/amrakk/zcago/model"
	"github.com/amrakk/zcago/session"
)

type (
	gifPayload struct {
		ClientID   string         `json:"clientId"`
		FileName   string         `json:"fileName"`
		TotalSize  int64          `json:"totalSize"`
		Width      int            `json:"width"`
		Height     int            `json:"height"`
		Msg        string         `json:"msg"`
		Type       int            `json:"type"`
		TTL        int            `json:"ttl"`
		Thumb      string         `json:"thumb"`
		Checksum   string         `json:"checksum"`
		TotalChunk int            `json:"totalChunk"`
		ChunkID    int            `json:"chunkId"`
		MetaData   map[string]any `json:"metaData,omitempty"`

		Visibility *int   `json:"visibility,omitempty"`
		Grid       string `json:"grid,omitempty"`
		ToID       string `json:"toid,omitempty"`
	}

	GIFContent struct {
		Attachment model.AttachmentSource
		Thumb      *model.AttachmentSource
		Msg        string
		TTL        int // Time to live in milliseconds
		Urgency    model.Urgency
	}

	SendGIFResponse struct {
		MsgID string `json:"msgId"`
		Href  string `json:"href"`
	}
	SendGIFFn = func(ctx context.Context, threadID string, threadType model.ThreadType, gif GIFContent) (*SendGIFResponse, error)
)

func (a *api) SendGIF(ctx context.Context, threadID string, threadType model.ThreadType, gif GIFContent) (*SendGIFResponse, error) {
	return a.e.SendGIF(ctx, threadID, threadType, gif)
}

var sendGIFFactory = apiFactory[*SendGIFResponse, SendGIFFn]()(
	func(a *api, sc session.Context, u factoryUtils[*SendGIFResponse]) (SendGIFFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("file"), "")
		shareFile := sc.Settings().Features.ShareFile
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/gif", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/gif", nil, true),
		}

		isExceedMaxFileSize := func(totalSize int64) bool {
			return totalSize > shareFile.MaxSizeShareFileV3*1024*1024
		}

		return func(ctx context.Context, threadID string, threadType model.ThreadType, content GIFContent) (*SendGIFResponse, error) {
			var (
				reader       io.Reader
				closer       io.Closer
				fileName     string
				fileMetadata model.AttachmentMetadata
			)

			if f := content.Attachment.String(); f != "" {
				r, err := os.Open(f)
				if err != nil {
					return nil, errs.WrapZCA("failed to read file", "api.SendGIF", err)
				}

				reader = r
				closer = r

				fileMetadata, fileName, err = sc.GetImageMetadata(f)
				if err != nil {
					return nil, err
				}
			} else if f := content.Attachment.Object(); f != nil {
				reader, fileName, fileMetadata = f.Data, f.Filename, f.Metadata
			}
			if closer != nil {
				defer closer.Close()
			}
			if reader == nil || fileName == "" {
				return nil, errs.ErrSourceEmpty
			}
			if isExceedMaxFileSize(fileMetadata.Size) {
				return nil, errs.ErrExceedMaxFileSize
			}

			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, errs.WrapZCA("failed to read attachment data", "api.SendGIF", err)
			}
			checksum := md5.Sum(data)

			if content.Thumb == nil {
				g, err := gif.DecodeAll(bytes.NewReader(data))
				if err != nil {
					return nil, errs.WrapZCA("failed to decode gif for thumbnail", "api.SendGIF", err)
				}

				first := g.Image[0]
				b := first.Bounds()

				var buf bytes.Buffer
				if err := png.Encode(&buf, first); err != nil {
					return nil, errs.WrapZCA("failed to encode png thumbnail", "api.SendGIF", err)
				}
				meta := model.AttachmentMetadata{
					Size:   int64(buf.Len()),
					Width:  b.Dx(),
					Height: b.Dy(),
				}

				thumbSource, err := model.NewObjectAttachment(
					fileName,
					meta,
					bytes.NewReader(buf.Bytes()),
				)
				if err != nil {
					return nil, err
				}

				content.Thumb = thumbSource
			}

			forms, err := httpx.BuildFormData(
				"chunkContent", bytes.NewReader(data),
				httpx.WithContentType("application/octet-stream"),
				httpx.WithFileName(fileName),
			)
			if err != nil || len(forms) == 0 || forms[0] == nil {
				return nil, errs.WrapZCA("failed to build form data", "api.SendGIF", err)
			}

			thumb, err := a.UploadThumbnail(ctx, *content.Thumb)
			if err != nil {
				return nil, errs.WrapZCA("failed to upload thumbnail", "api.SendGIF", err)
			}

			payload := gifPayload{
				ClientID:   strconv.FormatInt(time.Now().UnixMilli(), 10),
				FileName:   fileName,
				TotalSize:  fileMetadata.Size,
				Width:      fileMetadata.Width,
				Height:     fileMetadata.Height,
				Msg:        content.Msg,
				Type:       1,
				TTL:        content.TTL,
				Thumb:      thumb.URL,
				Checksum:   hex.EncodeToString(checksum[:]),
				TotalChunk: 1,
				ChunkID:    1,
			}
			if content.Urgency == model.UrgImportant || content.Urgency == model.UrgUrgent {
				payload.MetaData = map[string]any{"urgency": content.Urgency}
			}

			if threadType == model.ThreadTypeGroup {
				v := 0
				payload.Visibility = &v
				payload.Grid = threadID
			} else {
				payload.ToID = threadID
			}

			enc, err := u.EncodeAES(jsonx.Stringify(payload))
			if err != nil {
				return nil, errs.WrapZCA("failed to encrypt params", "api.SendGIF", err)
			}
			url := u.MakeURL(
				serviceURLs[threadType],
				map[string]any{"type": "1", "params": enc},
				true,
			)
			resp, err := u.Request(ctx, url, &httpx.RequestOptions{
				Method:  http.MethodPost,
				Headers: forms[0].Header,
				Body:    forms[0].Body,
			})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return resolveResponse[*SendGIFResponse](sc, resp, true)
		}, nil
	},
)

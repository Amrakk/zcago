package api

import (
	"bytes"
	"context"
	"errors"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Amrakk/zcago/config"
	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/model"
	"github.com/Amrakk/zcago/session"
	"golang.org/x/sync/errgroup"
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
		TTL        int // Time to live in milliseconds
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
		serviceURLs := map[model.ThreadType]string{
			model.ThreadTypeUser:  u.MakeURL(base+"/api/message/gif", nil, true),
			model.ThreadTypeGroup: u.MakeURL(base+"/api/group/gif", nil, true),
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

			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, errs.WrapZCA("failed to read attachment data", "api.SendGIF", err)
			}
			if closer != nil {
				_ = closer.Close()
			}

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
				httpx.WithChunkSize(config.GIFChunkSize),
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
				Msg:        "",
				Type:       1,
				TTL:        content.TTL,
				Thumb:      thumb.URL,
				Checksum:   content.Attachment.GetLargeFileMD5().Checksum,
				TotalChunk: len(forms),
				ChunkID:    1,
			}

			if threadType == model.ThreadTypeGroup {
				v := 0
				payload.Visibility = &v
				payload.Grid = threadID
			} else {
				payload.ToID = threadID
			}

			var results atomic.Pointer[SendGIFResponse]

			g, gctx := errgroup.WithContext(ctx)
			for i := range forms {
				chunk := i

				g.Go(func() error {
					p := payload
					p.ChunkID = chunk + 1

					enc, err := u.EncodeAES(jsonx.Stringify(p))
					if err != nil {
						return errs.WrapZCA("failed to encrypt params", "api.SendGIF", err)
					}
					url := u.MakeURL(
						serviceURLs[threadType],
						map[string]any{"type": "1", "params": enc},
						true,
					)

					reqCtx := gctx
					resp, err := u.Request(reqCtx, url, &httpx.RequestOptions{
						Method:  http.MethodPost,
						Headers: forms[chunk].Header,
						Body:    forms[chunk].Body,
					})
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					r, err := resolveResponse[*SendGIFResponse](sc, resp, true)
					if err != nil {
						var zerr errs.ZaloAPIError
						if errors.As(err, &zerr) && zerr.Code != nil && *zerr.Code == 220 {
							return nil
						}
						return err
					}
					results.Store(r)

					return nil
				})
			}
			if err := g.Wait(); err != nil {
				return nil, err
			}

			return results.Load(), nil
		}, nil
	},
)

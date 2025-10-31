package api

import (
	"context"
	"errors"
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

		return func(ctx context.Context, threadID string, threadType model.ThreadType, gif GIFContent) (*SendGIFResponse, error) {
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

			if f := gif.Attachment.String(); f != "" {
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
			} else if f := gif.Attachment.Object(); f != nil {
				reader, fileName, fileMetadata = f.Data, f.Filename, f.Metadata
			}

			forms, err := httpx.BuildFormData(
				"chunkContent", reader,
				httpx.WithContentType("application/octet-stream"),
				httpx.WithFileName(fileName),
				httpx.WithChunkSize(config.GIFChunkSize),
			)
			if err != nil || len(forms) == 0 || forms[0] == nil {
				return nil, errs.WrapZCA("failed to build form data", "api.SendGIF", err)
			}

			thumb, err := a.UploadThumbnail(ctx, gif.Attachment)
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
				TTL:        gif.TTL,
				Thumb:      thumb.URL,
				Checksum:   gif.Attachment.GetLargeFileMD5().Checksum,
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

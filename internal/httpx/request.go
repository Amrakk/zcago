package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Amrakk/zcago/session"
)

type RequestOptions struct {
	Method  string
	Headers http.Header
	Query   url.Values
	Body    io.Reader
	Raw     bool
}

func BuildFormBody(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	return strings.NewReader(form.Encode())
}

type FormData struct {
	Body   *bytes.Buffer
	Header http.Header
}

type BuildOptions struct {
	FileName    string // default: "blob"
	ContentType string // default: auto-detect then fallback to octet-stream
	ChunkSize   int64  // >0 => build chunks; else single form
}

type Opt func(*BuildOptions)

func WithFileName(name string) Opt  { return func(o *BuildOptions) { o.FileName = name } }
func WithContentType(ct string) Opt { return func(o *BuildOptions) { o.ContentType = ct } }
func WithChunkSize(n int64) Opt     { return func(o *BuildOptions) { o.ChunkSize = n } }

func BuildFormData(fieldName string, source io.Reader, opts ...Opt) ([]*FormData, error) {
	o := BuildOptions{FileName: "blob"}
	for _, f := range opts {
		f(&o)
	}

	ct, src := detectContentType(o.FileName, o.ContentType, source)
	if o.ChunkSize > 0 {
		return buildChunks(fieldName, ct, o.FileName, src, o.ChunkSize)
	}
	fd, err := buildOne(fieldName, ct, o.FileName, src)
	if err != nil {
		return nil, err
	}
	return []*FormData{fd}, nil
}

func detectContentType(fileName, explicitCT string, src io.Reader) (string, io.Reader) {
	if explicitCT != "" {
		return explicitCT, src
	}
	if ext := filepath.Ext(fileName); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct, src
		}
	}
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	ct := http.DetectContentType(head[:n])
	if ct == "" {
		ct = "application/octet-stream"
	}
	return ct, io.MultiReader(bytes.NewReader(head[:n]), src)
}

func buildOne(fieldName, contentType, fileName string, src io.Reader) (*FormData, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
	h.Set("Content-Type", contentType)

	part, err := w.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, src); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	hdr := make(http.Header, 1)
	hdr.Set("Content-Type", w.FormDataContentType())
	return &FormData{Body: body, Header: hdr}, nil
}

func buildChunks(fieldName, contentType, fileName string, src io.Reader, chunkSize int64) ([]*FormData, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize must be > 0")
	}
	var out []*FormData
	r := src
	for {
		lr := &io.LimitedReader{R: r, N: chunkSize}
		fd, err := buildOne(fieldName, contentType, fileName, lr)
		if err != nil {
			return nil, err
		}

		consumed := chunkSize - lr.N
		if consumed == 0 {
			break
		}
		out = append(out, fd)
		if consumed < chunkSize {
			break
		}
	}
	return out, nil
}

func buildRequest(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions) (*http.Request, error) {
	headers := http.Header{}

	method := "GET"
	if opt != nil && opt.Method != "" {
		method = opt.Method
	}

	if opt != nil && !opt.Raw {
		def, err := getDefaultHeaders(sc)
		if err != nil {
			return nil, err
		}
		mergeHeaders(headers, def)
	}

	if opt != nil && opt.Headers != nil {
		mergeHeaders(headers, opt.Headers)
	}

	var body io.Reader
	if opt != nil {
		body = opt.Body
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	return req, nil
}

func getDefaultHeaders(sc session.MutableContext) (http.Header, error) {
	if sc.UserAgent() == "" {
		return nil, fmt.Errorf("user agent is not available")
	}

	h := make(http.Header, 8)
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Origin", "https://chat.zalo.me")
	h.Set("Referer", "https://chat.zalo.me/")
	h.Set("User-Agent", sc.UserAgent())
	return h, nil
}

func mergeHeaders(dst, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for k, vals := range src {
		for _, v := range vals {
			dst.Set(k, v)
		}
	}
}

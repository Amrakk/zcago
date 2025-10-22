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

func BuildFormData(fieldName, contentType, fileName string, source io.Reader) (*FormData, error) {
	if fieldName == "" {
		return nil, fmt.Errorf("fieldName is required")
	}
	if fileName == "" {
		fileName = "blob"
	}

	if contentType == "" {
		if ext := filepath.Ext(fileName); ext != "" {
			contentType = mime.TypeByExtension(ext)
		}
		if contentType == "" {
			head := make([]byte, 512)
			n, _ := io.ReadFull(source, head)
			contentType = http.DetectContentType(head[:n])
			source = io.MultiReader(bytes.NewReader(head[:n]), source)
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fieldHeader := make(textproto.MIMEHeader)
	fieldHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
	fieldHeader.Set("Content-Type", contentType)

	part, err := writer.CreatePart(fieldHeader)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, source); err != nil {
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	h := make(http.Header, 1)
	h.Set("Content-Type", writer.FormDataContentType())

	return &FormData{Body: body, Header: h}, nil
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

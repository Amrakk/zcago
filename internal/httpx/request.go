package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func buildRequest(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions) (*http.Request, error) {
	origin := getOrigin(urlStr)
	headers := http.Header{}

	method := "GET"
	if opt != nil && opt.Method != "" {
		method = opt.Method
	}

	if opt != nil && !opt.Raw {
		def, err := getDefaultHeaders(sc, origin)
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

func getDefaultHeaders(sc session.Context, origin string) (http.Header, error) {
	if origin == "" {
		origin = "https://chat.zalo.me"
	}
	if sc == nil || len(sc.Cookies()) == 0 {
		return nil, fmt.Errorf("cookie is not available")
	}
	if sc.UserAgent() == "" {
		return nil, fmt.Errorf("user agent is not available")
	}

	cookieStr := cookieString(sc.Cookies(origin))

	h := make(http.Header, 8)
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	h.Set("Cookie", cookieStr)
	h.Set("Origin", "https://chat.zalo.me")
	h.Set("Referer", "https://chat.zalo.me/")
	h.Set("User-Agent", sc.UserAgent())
	return h, nil
}

// TODO: remove this after implementing a custom cookie jar
func cookieString(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}
	var b strings.Builder
	for i, cookie := range cookies {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(cookie.Name)
		b.WriteByte('=')
		b.WriteString(cookie.Value)
	}
	return b.String()
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

func getOrigin(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host}).String()
}

package httpx

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/session"
)

func getDefaultHeaders(sc session.Context, origin string) (http.Header, error) {
	if origin == "" {
		origin = "https://chat.zalo.me"
	}
	if sc == nil || len(sc.Cookies()) == 0 {
		return nil, errs.NewZCAError("cookie is not available", "context", nil)
	}
	if sc.UserAgent() == "" {
		return nil, errs.NewZCAError("user agent is not available", "context", nil)
	}

	cookieStr, err := cookieString(sc.Cookies(), origin)
	if err != nil {
		return nil, err
	}

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

func cookieString(cookies []*http.Cookie, origin string) (string, error) {
	if len(cookies) == 0 {
		return "", nil
	}
	var b strings.Builder
	for i, c := range cookies {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(c.Name)
		b.WriteByte('=')
		b.WriteString(c.Value)
	}
	return b.String(), nil
}

// ---- Public API ----
const maxRedirects = 10

func Request(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions, raw bool) (*http.Response, error) {
	return requestWithRedirect(ctx, sc, urlStr, opt, raw, 0)
}

func requestWithRedirect(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions, raw bool, depth int) (*http.Response, error) {
	if depth > maxRedirects {
		return nil, errs.NewZCAError("too many redirects", "request", nil)
	}

	origin := originOf(urlStr)
	headers := http.Header{}
	if !raw {
		def, err := getDefaultHeaders(sc, origin)
		if err != nil {
			return nil, err
		}
		mergeHeader(headers, def)
	}
	if opt != nil && opt.Headers != nil {
		mergeHeader(headers, opt.Headers)
	}

	method := "GET"
	var body io.Reader
	if opt != nil {
		body = opt.Body
	}

	// Request
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	// Client with manual redirect handling
	client := buildClient(sc)
	defer restoreRedirectPolicy(client)
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		// Return the first response and let us handle Location manually
		return http.ErrUseLastResponse
	}
	client.Jar = sc.CookieJar()

	// Do
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Persist cookies from Set-Cookie (domain-aware), only when not raw
	if !raw {
		if err := persistSetCookies(sc.CookieJar(), resp, origin); err != nil {
			Logger(sc).Warn("failed to persist cookies")
		}
	}

	if loc := resp.Header.Get("Location"); loc != "" {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		nextURL := ResolveURL(urlStr, loc)
		nextOpt := &RequestOptions{
			Headers: func() http.Header {
				h := headers.Clone()
				if !raw {
					h.Set("Referer", "https://id.zalo.me/")
				}
				return h
			}(),
			Body: nil,
		}
		return requestWithRedirect(ctx, sc, nextURL, nextOpt, raw, depth+1)
	}

	return resp, nil
}

// ---- Helpers ----
func buildClient(sc session.MutableContext) *http.Client {
	if sc.Client() != nil {
		cp := *sc.Client()
		return &cp
	}

	return &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: nil, // set per-call
		Jar:           nil, // set per-call
	}
}

func restoreRedirectPolicy(c *http.Client) {
	// nothing to restore for the per-call client; kept for symmetry
}

func mergeHeader(dst, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for k, vals := range src {
		for _, v := range vals {
			dst.Set(k, v)
		}
	}
}

func persistSetCookies(jar http.CookieJar, resp *http.Response, origin string) error {
	if jar == nil || resp == nil {
		return nil
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return err
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		return nil
	}

	for _, c := range cookies {
		target := originURL
		if c.Domain != "" {
			host := strings.TrimPrefix(c.Domain, ".")
			if host != "" && !strings.EqualFold(host, "zalo.me") {
				target = &url.URL{Scheme: "https", Host: host, Path: "/"}
			}
		}
		jar.SetCookies(target, []*http.Cookie{c})
	}
	return nil
}

func originOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "https://xxxxxxxxx"
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host}).String()
}

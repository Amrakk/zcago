package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/session"
)

const maxRedirects = 10

func Request(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions) (*http.Response, error) {
	return requestWithRedirect(ctx, sc, urlStr, opt, 0)
}

func HandleZaloResponse[T any](sc session.Context, resp *http.Response, isEncrypted bool) *ZaloResponse[T] {
	return handleZaloResponse[T](sc, resp, isEncrypted)
}

func requestWithRedirect(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions, depth int) (*http.Response, error) {
	if depth > maxRedirects {
		logger.Log(sc).
			Warn("Too many redirects, aborting request").
			Debug("Max redirects exceeded:", maxRedirects)
		return nil, errs.NewZCAError("too many redirects", "request", nil)
	}

	req, err := buildRequest(ctx, sc, urlStr, opt)
	if err != nil {
		return nil, err
	}

	resp, err := executeRequest(sc, req, false)
	if err != nil {
		return nil, err
	}

	origin := getOrigin(urlStr)
	handleCookies(sc, resp, origin, opt)

	if loc := resp.Header.Get("Location"); loc != "" {
		logger.Log(sc).
			Debug("Following redirect to:", loc).
			Verbose("Redirect depth:", depth+1)

		func(b io.ReadCloser) {
			if _, err := io.Copy(io.Discard, b); err != nil {
				logger.Log(sc).Warn("Failed to discard response body:", err)
			}
			_ = b.Close()
		}(resp.Body)

		nextURL := ResolveURL(urlStr, loc)
		nextOpt := &RequestOptions{
			Method: http.MethodGet,
			Headers: func() http.Header {
				if opt != nil && !opt.Raw {
					h := req.Header.Clone()
					h.Set("Referer", "https://id.zalo.me/")
					return h
				}
				return nil
			}(),
			Body: nil,
		}
		return requestWithRedirect(ctx, sc, nextURL, nextOpt, depth+1)
	}

	return resp, nil
}

func executeRequest(sc session.MutableContext, req *http.Request, followRedirects bool) (*http.Response, error) {
	client := sc.Client()
	if client == nil {
		client = http.DefaultClient
	}

	httpClient := *client
	httpClient.Jar = sc.CookieJar()

	if !followRedirects {
		httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return httpClient.Do(req)
}

func handleCookies(sc session.MutableContext, resp *http.Response, origin string, opt *RequestOptions) {
	if opt != nil && !opt.Raw {
		if err := persistSetCookies(sc, resp, origin); err != nil {
			logger.Log(sc).Warn("Failed to persist cookies:", err)
		}
	}
}

func persistSetCookies(sc session.MutableContext, resp *http.Response, origin string) error {
	jar := sc.CookieJar()
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

	for _, cookie := range cookies {
		target := originURL
		if cookie.Domain != "" {
			host := strings.TrimPrefix(cookie.Domain, ".")
			if host != "" && !strings.EqualFold(host, "zalo.me") {
				target = &url.URL{Scheme: "https", Host: host, Path: "/"}
			}
		}
		jar.SetCookies(target, []*http.Cookie{cookie})
	}
	return nil
}

func handleZaloResponse[T any](sc session.Context, resp *http.Response, isEncrypted bool) *ZaloResponse[T] {
	out := &ZaloResponse[T]{}

	if !IsSuccess(resp) {
		out.Meta.Code = resp.StatusCode
		out.Meta.Message = "Request failed with status " + resp.Status
		return out
	}

	base, err := ParseBaseResponse(resp)
	if err != nil {
		logger.Log(sc).Error("Failed to parse response:", err)
		out.Meta.Message = "Failed to parse response data"
		return out
	}
	if base.ErrorCode != 0 {
		out.Meta.Code = base.ErrorCode
		out.Meta.Message = base.ErrorMessage
		return out
	}
	if base.Data == nil || *base.Data == "" {
		return out
	}

	var payloadBytes []byte

	if isEncrypted {
		key := sc.SecretKey().Bytes()
		if key == nil {
			logger.Log(sc).Error("Failed to decode secret key:", err)
			out.Meta.Message = "Failed to decode secret key"
			return out
		}

		plain, err := cryptox.DecodeAES(key, *base.Data)
		if err != nil {
			logger.Log(sc).Error("Failed to decrypt payload:", err)
			out.Meta.Message = "Failed to decrypt response data"
			return out
		}
		payloadBytes = plain
	} else {
		payloadBytes = []byte(*base.Data)
	}

	var decoded Response[T]
	if err := json.Unmarshal(payloadBytes, &decoded); err != nil {
		logger.Log(sc).Error("Failed to unmarshal payload:", err)
		out.Meta.Message = "Failed to parse response data"
		return out
	}

	out.Data = decoded.Data
	return out
}

package httpx

import (
	"context"
	"encoding/base64"
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

func Request(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions, raw bool) (*http.Response, error) {
	return requestWithRedirect(ctx, sc, urlStr, opt, raw, 0)
}

func HandleZaloResponse[T any](sc session.Context, resp *http.Response, isEncrypted bool) *ZaloResponse[T] {
	return handleZaloResponse[T](sc, resp, isEncrypted)
}

func requestWithRedirect(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions, raw bool, depth int) (*http.Response, error) {
	if depth > maxRedirects {
		logger.Log(sc).
			Warn("Too many redirects, aborting request").
			Debug("Max redirects exceeded:", maxRedirects)
		return nil, errs.NewZCAError("too many redirects", "request", nil)
	}

	req, err := buildRequest(ctx, sc, urlStr, opt, raw)
	if err != nil {
		return nil, err
	}

	resp, err := executeRequest(sc, req, false)
	if err != nil {
		return nil, err
	}

	origin := getOrigin(urlStr)
	if err := handleCookies(sc, resp, origin, raw); err != nil {
		return nil, err
	}

	if loc := resp.Header.Get("Location"); loc != "" {
		logger.Log(sc).
			Debug("Following redirect to:", loc).
			Verbose("Redirect depth:", depth+1)

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		nextURL := ResolveURL(urlStr, loc)
		nextOpt := &RequestOptions{
			Headers: func() http.Header {
				h := req.Header.Clone()
				if !raw {
				}
				return h
			}(),
			Body: nil,
		}
		return requestWithRedirect(ctx, sc, nextURL, nextOpt, raw, depth+1)
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

func handleCookies(sc session.MutableContext, resp *http.Response, origin string, raw bool) error {
	if !raw {
		if err := persistSetCookies(sc, resp, origin); err != nil {
			logger.Log(sc).Warn("Failed to persist cookies:", err)
		}
	}
	return nil
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
		key, err := base64.StdEncoding.DecodeString(sc.SecretKey())
		if err != nil {
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

	var decoded T
	if err := json.Unmarshal(payloadBytes, &decoded); err != nil {
		logger.Log(sc).Error("Failed to unmarshal payload:", err)
		out.Meta.Message = "Failed to parse response data"
		return out
	}

	out.Data = &decoded
	return out
}

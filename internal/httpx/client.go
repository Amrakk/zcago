package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/amrakk/zcago/config"
	"github.com/amrakk/zcago/internal/cryptox"
	"github.com/amrakk/zcago/internal/logger"
	"github.com/amrakk/zcago/session"
)

func Request(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions) (*http.Response, error) {
	return requestWithRedirect(ctx, sc, urlStr, opt, 0)
}

func HandleZaloResponse[T any](sc session.Context, resp *http.Response, isEncrypted bool) *ZaloResponse[T] {
	return handleZaloResponse[T](sc, resp, isEncrypted)
}

func requestWithRedirect(ctx context.Context, sc session.MutableContext, urlStr string, opt *RequestOptions, depth int) (*http.Response, error) {
	if depth > config.MaxRedirects {
		logger.Log(sc).
			Warn("Too many redirects, aborting request").
			Debug("Max redirects exceeded:", config.MaxRedirects)
		return nil, fmt.Errorf("too many redirects")
	}

	req, err := buildRequest(ctx, sc, urlStr, opt)
	if err != nil {
		return nil, err
	}

	resp, err := executeRequest(sc, req, false)
	if err != nil {
		return nil, err
	}

	if loc := resp.Header.Get("Location"); loc != "" {
		logger.Log(sc).
			Debug("Following redirect to: ", loc).
			Verbose("Redirect depth: ", depth+1)

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
	baseClient := sc.Client()
	if baseClient == nil {
		baseClient = http.DefaultClient
	}
	client := *baseClient
	if client.Jar == nil {
		client.Jar = sc.CookieJar()
	}

	if !followRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client.Do(req) //nolint:gosec // Generic HTTP helper intentionally sends URLs selected by its callers.
}

func handleZaloResponse[T any](
	sc session.Context,
	resp *http.Response,
	isEncrypted bool,
) *ZaloResponse[T] {
	out := &ZaloResponse[T]{}

	if !IsSuccess(resp) {
		return buildError[T](
			resp.StatusCode,
			"Request failed with status "+resp.Status,
		)
	}

	base, err := ParseBaseResponse(resp)
	if err != nil {
		logger.Log(sc).Error("Failed to parse response:", err)
		return buildError[T](0, "Failed to parse response data")
	}

	if base.ErrorCode != 0 {
		return buildError[T](base.ErrorCode, base.ErrorMessage)
	}

	if base.Data == nil || *base.Data == "" {
		return out
	}

	var payloadBytes []byte

	if isEncrypted {
		key := sc.SecretKey().Bytes()
		if len(key) == 0 {
			logger.Log(sc).Error("Secret key is empty")
			return buildError[T](0, "Failed to decode secret key")
		}

		plain, err := cryptox.DecodeAESCBC(key, *base.Data)
		if err != nil {
			logger.Log(sc).Error("Failed to decrypt payload:", err)
			return buildError[T](0, "Failed to decrypt response data")
		}

		payloadBytes = plain
	} else {
		payloadBytes = []byte(*base.Data)
	}

	var decodedMeta Response[json.RawMessage]
	if err := json.Unmarshal(payloadBytes, &decodedMeta); err != nil {
		logger.Log(sc).Error("Failed to unmarshal payload:", err)
		return buildError[T](0, "Failed to parse response data")
	}

	if decodedMeta.ErrorCode != 0 {
		return buildError[T](decodedMeta.ErrorCode, decodedMeta.ErrorMessage)
	}

	if len(decodedMeta.Data) == 0 || string(decodedMeta.Data) == "null" {
		return out
	}

	var decoded T
	if err := json.Unmarshal(decodedMeta.Data, &decoded); err != nil {
		logger.Log(sc).Error("unmarshal data field:", err)
		return buildError[T](0, "Failed to parse response data")
	}

	out.Data = decoded
	return out
}

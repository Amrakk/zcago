package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/internal/cryptox"
	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/listener"
	"github.com/Amrakk/zcago/session"
)

func New(sc session.MutableContext) (*api, error) {
	a := &api{sc: sc}

	if err := a.initEndpoints(); err != nil {
		return nil, err
	}

	l, err := listener.New(sc, sc.ZPWWebsocket())
	if err != nil {
		return nil, err
	}
	a.l = l

	return a, nil
}

type api struct {
	sc session.MutableContext
	e  endpoints
	l  listener.Listener
}

type endpoints struct {
	//gen:fields

	FetchAccountInfo FetchAccountInfoFn
	GetUserInfo      GetUserInfoFn
	UpdateLanguage   UpdateLanguageFn
}

func (a *api) initEndpoints() error {
	if !a.sc.SecretKey().IsValid() {
		return errs.NewZCAError("secret key missing or invalid", "", nil)
	}

	return firstErr(
		//gen:binds

		bind(a.sc, a, &a.e.FetchAccountInfo, fetchAccountInfoFactory),
		bind(a.sc, a, &a.e.GetUserInfo, getUserInfoFactory),
		bind(a.sc, a, &a.e.UpdateLanguage, updateLanguageFactory),
	)
}

type factoryUtils[T any] struct {
	MakeURL   func(baseURL string, params map[string]interface{}, includeDefaults bool) string
	EncodeAES func(data string) (string, error)
	Request   func(ctx context.Context, url string, options *httpx.RequestOptions) (*http.Response, error)
	Logger    *logger.Logger
	Resolve   func(res *http.Response, cb func(result *httpx.ZaloResponse[T]) T, isEncrypted bool) (T, error)
}

type (
	handler[T any, R any]         func(api *api, sc session.Context, utils factoryUtils[T]) (R, error)
	endpointFactory[T any, R any] func(sc session.MutableContext, api *api) (R, error)
)

func apiFactory[T any, R any]() func(
	callback handler[T, R],
) endpointFactory[T, R] {
	return func(callback handler[T, R]) endpointFactory[T, R] {
		return func(sc session.MutableContext, a *api) (R, error) {
			utils := factoryUtils[T]{
				MakeURL: func(url string, params map[string]any, includeDefaults bool) string {
					return httpx.MakeURL(sc, url, params, includeDefaults)
				},
				EncodeAES: func(data string) (string, error) {
					key := sc.SecretKey().Bytes()
					return cryptox.EncodeAESCBC(key, data, cryptox.EncryptTypeBase64)
				},
				Request: func(ctx context.Context, url string, opts *httpx.RequestOptions) (*http.Response, error) {
					return httpx.Request(ctx, sc, url, opts)
				},
				Logger: logger.Log(sc),
				Resolve: func(res *http.Response, cb func(*httpx.ZaloResponse[T]) T, isEncrypted bool) (T, error) {
					return resolveResponse(sc, res, cb, isEncrypted)
				},
			}

			return callback(a, sc, utils)
		}
	}
}

func resolveResponse[T any](
	sc session.Context,
	res *http.Response,
	cb func(*httpx.ZaloResponse[T]) T,
	isEncrypted bool,
) (T, error) {
	var zero T

	r := httpx.HandleZaloResponse[T](sc, res, isEncrypted)
	if r == nil {
		return zero, errs.NewZCAError("nil ZaloResponse", "resolveResponse", nil)
	}
	if r.Meta.Code != 0 {
		var zero T
		code := errs.ZaloErrorCode(r.Meta.Code)
		return zero, errs.NewZaloAPIError(r.Meta.Message, &code)
	}
	if cb != nil {
		return cb(r), nil
	}
	return r.Data, nil
}

func bind[T any, F any](sc session.MutableContext, a *api, target *F, factory endpointFactory[T, F]) error {
	fn, err := factory(sc, a)
	if err != nil {
		return err
	}
	*target = fn
	return nil
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

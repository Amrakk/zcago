package websocketx

import (
	"net/http"
	"net/url"
)

type Options struct {
	Proxy    func(*http.Request) (*url.URL, error)
	Header   http.Header
	MsgBuf   int
	ErrBuf   int
	WriteBuf int
}

func defaultOptions() Options {
	return Options{
		MsgBuf:   256,
		ErrBuf:   8,
		WriteBuf: 64,
	}
}

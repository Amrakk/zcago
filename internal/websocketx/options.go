package websocketx

import (
	"net/http"
)

type Options struct {
	Header     http.Header
	HTTPClient *http.Client
	MsgBuf     int
	ErrBuf     int
	WriteBuf   int
}

func defaultOptions() Options {
	return Options{
		MsgBuf:   64,
		ErrBuf:   8,
		WriteBuf: 16,
	}
}

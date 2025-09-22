package httpx

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
)

type BaseResponse[T any] struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         *T     `json:"data"`
}

type EncryptedResponse = BaseResponse[string]

type ZaloResponse[T any] struct {
	Meta struct {
		Code    int
		Message string
	}
	Data *T
}

func DecodeBody(resp *http.Response) (io.ReadCloser, error) {
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		return gzip.NewReader(resp.Body)
	case "deflate":
		return zlib.NewReader(resp.Body)
	// case "br":
	// case "zstd":
	default:
		return resp.Body, nil
	}
}

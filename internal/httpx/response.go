package httpx

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
)

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

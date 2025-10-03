package httpx

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Amrakk/zcago/internal/errs"
)

func DecodeResponse(resp *http.Response) (io.ReadCloser, error) {
	if resp == nil {
		return nil, errs.NewZCAError("response is nil", "DecodeResponse", nil)
	}

	contentEncoding := resp.Header.Get("Content-Encoding")
	switch strings.ToLower(contentEncoding) {
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

func ReadJSON(resp *http.Response, target interface{}) error {
	body, err := DecodeResponse(resp)
	if err != nil {
		return err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)
	if err := decoder.Decode(target); err != nil {
		return errs.NewZCAError("failed to decode JSON", "ReadJSON", &err)
	}

	return nil
}

func ReadBytes(resp *http.Response) ([]byte, error) {
	body, err := DecodeResponse(resp)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(body)
}

func ReadString(resp *http.Response) (string, error) {
	bytes, err := ReadBytes(resp)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func IsSuccess(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func IsClientError(resp *http.Response) bool {
	return resp.StatusCode >= 400 && resp.StatusCode < 500
}

func IsServerError(resp *http.Response) bool {
	return resp.StatusCode >= 500 && resp.StatusCode < 600
}

func CheckStatus(resp *http.Response) error {
	if IsSuccess(resp) {
		return nil
	}

	body, _ := ReadString(resp)
	return errs.NewZCAError(fmt.Sprintf("Status %d: %s", resp.StatusCode, body), "", nil)
}

type Response[T any] struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         T      `json:"data"`
}

type BaseResponse = Response[*string]

type ZaloResponse[T any] struct {
	Meta struct {
		Code    int
		Message string
	}
	Data T
}

func ParseBaseResponse(resp *http.Response) (*BaseResponse, error) {
	var result BaseResponse
	if err := ReadJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func ParseZaloResponse[T any](resp *http.Response) (*ZaloResponse[T], error) {
	var result ZaloResponse[T]
	if err := ReadJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

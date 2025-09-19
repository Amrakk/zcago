package errs

import (
	"fmt"
	"strings"
)

type ZaloErrorCode int

const (
	ZaloErrorCodeInvalidParams ZaloErrorCode = 114
)

type ZaloAPIError struct {
	Code *ZaloErrorCode
	Msg  string
}

func NewZaloAPIError(msg string, code *ZaloErrorCode) *ZaloAPIError {
	return &ZaloAPIError{Code: code, Msg: msg}
}

func (e *ZaloAPIError) Error() string {
	errName := "ZaloAPIError"
	parts := []string{errName}

	if e.Msg != "" {
		parts = append(parts, e.Msg)
	}
	if e.Code != nil {
		parts = append(parts, fmt.Sprintf("[%d]", *e.Code))
	}

	return strings.Join(parts, ": ")
}

package errs

import "fmt"

type ZaloErrorCode int

const (
	ZaloErrorCodeInvalidParams ZaloErrorCode = 114
)

type ZaloAPIError struct {
	Code    *ZaloErrorCode
	Message string
}

func (e *ZaloAPIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != nil {
		return fmt.Sprintf("ZaloAPIError[%d]: %s", *e.Code, e.Message)
	}
	return fmt.Sprintf("ZaloAPIError: %s", e.Message)
}

func NewZaloAPIError(msg string, code *ZaloErrorCode) *ZaloAPIError {
	return &ZaloAPIError{
		Code:    code,
		Message: msg,
	}
}

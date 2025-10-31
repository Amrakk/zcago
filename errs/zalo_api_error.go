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

func (e ZaloAPIError) Error() string {
	if e.Code != nil {
		return fmt.Sprintf("ZaloAPIError[%d]: %s", *e.Code, e.Message)
	}
	return "ZaloAPIError: " + e.Message
}

func (e ZaloAPIError) Is(target error) bool {
	if target, ok := target.(ZaloAPIError); ok {
		if e.Code == nil && target.Code == nil {
			return e.Message == target.Message
		}
		if e.Code != nil && target.Code != nil {
			return *e.Code == *target.Code && e.Message == target.Message
		}
		return false
	}
	return false
}

func NewZaloAPIError(msg string, code *ZaloErrorCode) ZaloAPIError {
	return ZaloAPIError{
		Code:    code,
		Message: msg,
	}
}

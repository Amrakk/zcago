package errs

import (
	"fmt"
)

var (
	ErrLoginQRAborted  = NewZCA("login QR aborted by user", "")
	ErrLoginQRDeclined = NewZCA("login QR declined by user", "")

	ErrMissingImageMetadataGetter = NewZCA("missing `imageMetadataGetter`. Please provide it in the Zalo object options.", "")
)

type ZCAError struct {
	Message string
	Op      string
	Cause   error
	Meta    map[string]string
}

func (e ZCAError) Error() string {
	base := "ZCAError"

	switch {
	case e.Message != "" && e.Op != "":
		return fmt.Sprintf("%s: %s (%s)", base, e.Message, e.Op)
	case e.Message != "":
		return fmt.Sprintf("%s: %s", base, e.Message)
	case e.Op != "":
		return fmt.Sprintf("%s: (%s)", base, e.Op)
	default:
		return base
	}
}

func (e ZCAError) Unwrap() error { return e.Cause }

func (e ZCAError) Is(target error) bool {
	if target, ok := target.(ZCAError); ok {
		return e.Message == target.Message && e.Op == target.Op
	}
	return false
}

func NewZCA(msg, op string) ZCAError { return ZCAError{Message: msg, Op: op} }

func WrapZCA(msg, op string, cause error) ZCAError {
	return ZCAError{Message: msg, Op: op, Cause: cause}
}

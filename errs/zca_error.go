package errs

import (
	"fmt"

	"github.com/Amrakk/zcago/config"
)

var (
	ErrLoginQRAborted  = NewZCA("login QR aborted by user", "")
	ErrLoginQRDeclined = NewZCA("login QR declined by user", "")

	ErrMissingImageMetadataGetter = NewZCA("missing `imageMetadataGetter`. Please provide it in the Zalo object options.", "")

	ErrFileContentUnavailable     = NewZCA("unable to get file content", "")
	ErrInvalidMessageCount        = NewZCA(fmt.Sprintf("message count out of range (allowed: 1-%d)", config.MaxMessagesPerRequest), "")
	ErrInconsistentGroupRecipient = NewZCA("all messages in a group thread must share the same idTo", "")

	ErrSourceEmpty       = NewZCA("source cannot be empty", "")
	ErrExceedMaxFile     = NewZCA("exceeded maximum number of files per request", "")
	ErrInvalidExtension  = NewZCA("file has an invalid extension", "")
	ErrExceedMaxFileSize = NewZCA("exceeded maximum file size", "")
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

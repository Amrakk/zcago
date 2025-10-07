package errs

import "fmt"

type ZCAError struct {
	Message string
	Op      string
	Cause   error
	Meta    map[string]string
}

func (e *ZCAError) Error() string {
	if e == nil {
		return "<nil>"
	}

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

func (e *ZCAError) Unwrap() error { return e.Cause }

func NewZCA(msg, op string) *ZCAError { return &ZCAError{Message: msg, Op: op} }

func WrapZCA(msg, op string, cause error) *ZCAError {
	return &ZCAError{Message: msg, Op: op, Cause: cause}
}

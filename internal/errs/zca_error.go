package errs

import (
	"fmt"
	"strings"
)

type ZCAErrorCode int

const (
	ZCAErrorCodeInvalidParams ZCAErrorCode = 114
)

type ZCAError struct {
	Msg string
	Op  string
	Err *error
}

func NewZCAError(msg string, op string, err *error) *ZCAError {
	return &ZCAError{Msg: msg, Op: op, Err: err}
}

func (e *ZCAError) Error() string {
	errName := "ZCAError"

	parts := []string{errName}

	if e.Msg != "" {
		parts = append(parts, e.Msg)
	}
	if e.Op != "" {
		parts = append(parts, fmt.Sprintf("op=%s", e.Op))
	}

	return strings.Join(parts, ": ")
}

func (e *ZCAError) Unwrap() error { return *e.Err }

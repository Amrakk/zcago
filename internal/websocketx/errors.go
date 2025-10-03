package websocketx

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
)

func isCloseErr(err error) bool {
	var ce *websocket.CloseError
	return errors.As(err, &ce)
}

func isEOForUnexpected(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

func isConnClosedOrReset(err error) bool {
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	var op *net.OpError
	if errors.As(err, &op) {
		if se, ok := op.Err.(syscall.Errno); ok {
			if se == syscall.EPIPE || se == syscall.ECONNRESET {
				return true
			}
		}
	}

	return strings.Contains(err.Error(), "use of closed network connection")
}

func isFatalWriteErr(err error) bool {
	return isCloseErr(err) || isEOForUnexpected(err) || isConnClosedOrReset(err)
}

func closeInfoFromErr(err error) CloseInfo {
	var ci CloseInfo
	var ce *websocket.CloseError
	if errors.As(err, &ce) {
		ci.Code = ce.Code
		ci.Reason = ce.Text
	} else {
		ci.Err = err
	}
	return ci
}

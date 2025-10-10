package websocketx

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/coder/websocket"
)

func (c *client) isFatalErr(err error) bool {
	if c.isFatalContextErr(err) {
		return true
	}

	if isContextErr(err) {
		return false
	}

	return isCloseErr(err) || isEOForUnexpected(err) || isConnClosedOrReset(err)
}

func (c *client) isFatalContextErr(err error) bool {
	if !isContextErr(err) {
		return false
	}

	return c.connCtx.Err() != nil
}

func closeInfoFromErr(err error) CloseInfo {
	var ci CloseInfo
	var ce websocket.CloseError
	if errors.As(err, &ce) {
		ci.Code = int(ce.Code)
		ci.Reason = ce.Reason
	} else {
		ci.Code = int(websocket.StatusInternalError)
		ci.Err = err
	}
	return ci
}

func isCloseErr(err error) bool {
	var ce websocket.CloseError
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

func isContextErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

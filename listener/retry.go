package listener

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/internal/websocketx"
)

type retryState struct {
	count int
	max   int
	times []int
}

func (ln *listener) shouldRetryConnection(ctx context.Context, ci websocketx.CloseInfo, retryOnClose bool) (int, bool) {
	if !retryOnClose || ctx.Err() != nil {
		return 0, false
	}

	return ln.canRetry(ci.Code)
}

func (ln *listener) scheduleReconnection(ctx context.Context, ci websocketx.CloseInfo, delay int) error {
	if ln.shouldRotate(ci.Code) {
		if err := ln.rotateEndpoint(); err != nil {
			return err
		}
	}

	time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := ln.Start(ctx, true); err != nil {
			select {
			case ln.ch.Closed <- ci:
			default:
			}
		}
	})

	return nil
}

func (ln *listener) canRetry(code int) (int, bool) {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if !ln.shouldRetry(code) {
		return 0, false
	}

	st, ok := ln.retryStates[fmt.Sprint(code)]
	if !ok || st == nil || st.max == 0 || len(st.times) == 0 {
		return 0, false
	}

	if st.count >= st.max {
		return 0, false
	}

	idx := st.count
	st.count++

	var delay int
	if idx < len(st.times) {
		delay = st.times[idx]
	} else {
		delay = st.times[len(st.times)-1]
	}

	logger.Log(ln.sc).Verbosef(
		"Retry for code %d in %dms (%d/%d)",
		code, delay, st.count, st.max,
	)

	return delay, true
}

func (ln *listener) shouldRetry(code int) bool {
	if ln.retryStates != nil {
		_, ok := ln.retryStates[fmt.Sprint(code)]
		return ok
	}

	s := ln.sc.Settings()
	if s == nil || s.Features.Socket.CloseAndRetry == nil {
		return false
	}

	return slices.Contains(s.Features.Socket.CloseAndRetry, code)
}

func (ln *listener) shouldRotate(code int) bool {
	s := ln.sc.Settings()
	if s == nil || s.Features.Socket.RotateErrorCodes == nil {
		return false
	}

	if slices.Contains(s.Features.Socket.RotateErrorCodes, code) {
		return ln.rotateCount < len(ln.urls)-1
	}

	return false
}

func (ln *listener) rotateEndpoint() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	ln.rotateCount++
	wsURL := httpx.MakeURL(ln.sc, ln.urls[ln.rotateCount], map[string]any{
		"t": time.Now().UnixMilli(),
	}, true)
	if wsURL == "" {
		return errs.NewZCA("build websocket URL failed", "listener.rotateEndpoint")
	}
	ln.wsURL = wsURL
	logger.Log(ln.sc).Verbosef(`Rotating endpoint to %s`, ln.wsURL)

	return nil
}

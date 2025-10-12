package timex

import (
	"context"
	"time"
)

func SetTimeout(ctx context.Context, d time.Duration, fn func()) func() {
	t := time.NewTimer(d)
	done := make(chan struct{})

	childCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(done)
		select {
		case <-t.C:
			if childCtx.Err() == nil {
				fn()
			}
		case <-childCtx.Done():
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
			return
		}
	}()

	return func() {
		cancel()
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		<-done
	}
}

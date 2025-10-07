package listener

import (
	"context"
	"fmt"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/listener/events"
	"github.com/Amrakk/zcago/model"
)

func (ln *listener) router(version, cmd, sub uint, body BaseWSMessage) {
	key := fmt.Sprintf("%d_%d_%d", version, cmd, sub)

	switch key {
	case "1_1_1":
		ln.handleCipherKey(body)

	case "1_501_0":
		ln.handleMessages(body)

	case "1_3000_0":
		ln.handleDuplicateConnection()

	default:
	}
}

func (ln *listener) handleCipherKey(body BaseWSMessage) {
	key := *body.Key
	if key == "" {
		return
	}

	ln.cipherKey = key
	ln.ch.CipherKey <- key

	if ln.pingStopper != nil {
		(*ln.pingStopper)()
	}

	ping := func() {
		payload := WSPayload{
			Version: 1,
			CMD:     2,
			SubCMD:  1,
			Data:    map[string]any{"eventId": time.Now().UnixMilli()},
		}
		if err := ln.SendWS(ln.ctx, payload, false); err != nil {
			ln.emitError(ln.ctx, errs.WrapZCA("failed to send ping:", "ping", err))
		}
	}

	interval := ln.sc.WSPingInterval()
	if interval <= 0 {
		return
	}

	stop := startPingLoop(ln.ctx, interval, ping)
	ln.pingStopper = &stop
}

func (ln *listener) handleMessages(body BaseWSMessage) {
	eventData, err := decodeEventData[events.MessageEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleMessages", err)
		ln.emitError(ln.ctx, err)
		return
	}

	for _, msg := range eventData.Data.Msgs {
		if msg.Undo != nil {
			undoObject := model.NewUndo(ln.sc.UID(), *msg.Undo, false)
			if undoObject.IsSelf && !ln.selfListen {
				continue
			}
			emit(ln.ch.Undo, undoObject)
		} else if msg.Message != nil {
			messageObject := model.NewUserMessage(ln.sc.UID(), *msg.Message)
			if messageObject.IsSelf && !ln.selfListen {
				continue
			}
			emit(ln.ch.Message, messageObject)
		}
	}
}

func (ln *listener) handleDuplicateConnection() {
	logger.Log(ln.sc).Error()
	logger.Log(ln.sc).Error("Another connection is opened, closing this one")
	logger.Log(ln.sc).Error()
	ln.client.Close(ZaloDuplicateConnection, "Another connection is opened, closing this one")
}

//
// Helpers
//

func startPingLoop(ctx context.Context, interval time.Duration, f func()) func() {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()
		f()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				f()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() { close(done) }
}

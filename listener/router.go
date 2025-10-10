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

func (ln *listener) router(ctx context.Context, version, cmd, sub uint, body BaseWSMessage) {
	key := fmt.Sprintf("%d_%d_%d", version, cmd, sub)

	switch key {
	case "1_1_1":
		ln.handleCipherKey(ctx, body)

	case "1_501_0":
		ln.handleMessages(ctx, body)

	case "1_3000_0":
		ln.handleDuplicateConnection()

	default:
	}
}

func (ln *listener) handleCipherKey(ctx context.Context, body BaseWSMessage) {
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
		if err := ln.SendWS(ctx, payload, false); err != nil {
			ln.emitError(ctx, errs.WrapZCA("failed to send ping:", "ping", err))
		}
	}

	interval := ln.sc.WSPingInterval()
	if interval <= 0 {
		return
	}

	stop := startPingLoop(ctx, interval, ping)
	ln.pingStopper = &stop
}

func (ln *listener) handleMessages(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.MessageEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleMessages", err)
		ln.emitError(ctx, err)
		return
	}

	for _, msg := range eventData.Data.Msgs {
		if msg.Undo != nil {
			undoObject := model.NewUndo(ln.sc.UID(), *msg.Undo, false)
			if undoObject.IsSelf && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Undo, undoObject)
		} else if msg.Message != nil {
			messageObject := model.NewUserMessage(ln.sc.UID(), *msg.Message)
			if messageObject.IsSelf && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Message, messageObject)
		}
	}
}

func (ln *listener) handleDuplicateConnection() {
	logger.Log(ln.sc).Error()
	logger.Log(ln.sc).Error("Another connection is opened, closing this one")
	logger.Log(ln.sc).Error()

	client := ln.getClient()
	if client != nil {
		client.Close(ZaloDuplicateConnection, "Another connection is opened, closing this one")
	}
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
				return
			}
		}
	}()

	return func() { close(done) }
}

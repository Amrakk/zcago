package listener

import (
	"context"
	"time"

	"github.com/Amrakk/zcago/internal/errs"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/listener/events"
	"github.com/Amrakk/zcago/model"
)

func (ln *listener) router(version, cmd, sub uint, body BaseWSMessage) {
	switch {
	case version == 1 && cmd == 1 && sub == 1:
		ln.handleCipherKey(body)

	case version == 1 && cmd == 501 && sub == 0:
		ln.handleMessages(body)

	case version == 1 && cmd == 3000 && sub == 0:
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
		ln.SendWS(ln.ctx, payload, false)
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
		err = errs.NewZCAError("Failed to decode event data:", "handleMessages", &err)
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
	ln.client.Close(ZaloDuplicateConnection, "")
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

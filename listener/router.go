package listener

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Amrakk/zcago/errs"
	"github.com/Amrakk/zcago/internal/logger"
	"github.com/Amrakk/zcago/listener/events"
	"github.com/Amrakk/zcago/model"
)

func (ln *listener) router(ctx context.Context, version, cmd, sub uint, body BaseWSMessage) {
	key := fmt.Sprintf("%d_%d_%d", version, cmd, sub)

	fmt.Println(key)

	switch key {
	case "1_1_1":
		ln.handleCipherKey(ctx, body)

	case "1_501_0":
		ln.handleMessages(ctx, body)

	case "1_502_0":
		ln.handleMessagesStatus(ctx, body)

	case "1_510_1", "1_511_1":
		ln.handleOldMessages(ctx, body)

	case "1_521_0":
		ln.handleGroupMessages(ctx, body)

	case "1_522_0":
		ln.handleGroupMessagesStatus(ctx, body)

	case "1_601_0":
		ln.handleControls(ctx, body)

	case "1_602_0":
		ln.handleActions(ctx, body)

	case "1_610_1", "1_611_1":
		ln.handleOldReactions(ctx, body)

	case "1_612_0":
		ln.handleReactions(ctx, body)

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

	uid := ln.sc.UID()
	for _, msg := range eventData.Data.Msgs {
		if msg.Undo != nil {
			undo := model.NewUndo(uid, *msg.Undo, false)
			if undo.IsSelf && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Undo, undo)
		} else if msg.Message != nil {
			message := model.NewUserMessage(uid, *msg.Message)
			if message.IsSelf() && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Message, model.Message(message))
		}
	}
}

func (ln *listener) handleOldMessages(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.OldMessagesEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleOldMessages", err)
		ln.emitError(ctx, err)
		return
	}

	messages := make([]model.Message, 0, len(eventData.Data.Msgs)+len(eventData.Data.GroupMsgs))

	threadType := model.ThreadTypeUser
	selected := eventData.Data.Msgs
	if len(eventData.Data.GroupMsgs) > 0 {
		threadType = model.ThreadTypeGroup
		selected = eventData.Data.GroupMsgs
	}

	uid := ln.sc.UID()
	for _, msg := range selected {
		messageObject := model.NewUserMessage(uid, msg)
		messages = append(messages, messageObject)
	}

	emit(ctx, ln.ch.OldMessages, model.NewOldMessage(messages, threadType))
}

func (ln *listener) handleMessagesStatus(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.MessageStatusEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleMessagesStatus", err)
		ln.emitError(ctx, err)
		return
	}

	// Always triggered by others; no self-check needed
	if len(eventData.Data.DeliveredMessages) > 0 {
		deliveredMsgs := make([]model.DeliveredMessage, 0, len(eventData.Data.DeliveredMessages))
		for _, dm := range eventData.Data.DeliveredMessages {
			deliveredMsgs = append(deliveredMsgs, model.NewUserDeliveredMessage(dm))
		}
		emit(ctx, ln.ch.DeliveredMessages, deliveredMsgs)
	}
	if len(eventData.Data.SeenMessages) > 0 {
		seenMsgs := make([]model.SeenMessage, 0, len(eventData.Data.SeenMessages))
		for _, sm := range eventData.Data.SeenMessages {
			seenMsgs = append(seenMsgs, model.NewUserSeenMessage(sm))
		}
		emit(ctx, ln.ch.SeenMessages, seenMsgs)
	}
}

func (ln *listener) handleGroupMessages(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.GroupMessageEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleGroupMessages", err)
		ln.emitError(ctx, err)
		return
	}

	for _, msg := range eventData.Data.GroupMsgs {
		if msg.Undo != nil {
			undo := model.NewUndo(ln.sc.UID(), *msg.Undo, true)
			if undo.IsSelf && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Undo, undo)
		} else if msg.Message != nil {
			message := model.NewGroupMessage(ln.sc.UID(), *msg.Message)
			if message.IsSelf() && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Message, model.Message(message))
		}
	}
}

func (ln *listener) handleGroupMessagesStatus(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.GroupMessageStatusEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleGroupMessagesStatus", err)
		ln.emitError(ctx, err)
		return
	}

	uid := ln.sc.UID()
	if len(eventData.Data.DeliveredMessages) > 0 {
		deliveredMsgs := make([]model.DeliveredMessage, 0, len(eventData.Data.DeliveredMessages))
		for _, dm := range eventData.Data.DeliveredMessages {
			deliveredObject := model.NewGroupDeliveredMessage(uid, dm)
			if deliveredObject.IsSelf() && !ln.selfListen {
				continue
			}
			deliveredMsgs = append(deliveredMsgs, deliveredObject)

		}
		emit(ctx, ln.ch.DeliveredMessages, deliveredMsgs)
	}
	if len(eventData.Data.SeenMessages) > 0 {
		seenMsgs := make([]model.SeenMessage, 0, len(eventData.Data.SeenMessages))
		for _, sm := range eventData.Data.SeenMessages {
			seenObject := model.NewGroupSeenMessage(uid, sm)
			if seenObject.IsSelf() && !ln.selfListen {
				continue
			}
			seenMsgs = append(seenMsgs, seenObject)
		}
		emit(ctx, ln.ch.SeenMessages, seenMsgs)
	}
}

func (ln *listener) handleReactions(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.ReactionEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleReaction", err)
		ln.emitError(ctx, err)
		return
	}

	uid := ln.sc.UID()
	for _, r := range eventData.Data.Reactions {
		reaction := model.NewReaction(uid, r, model.ThreadTypeUser)
		if reaction.IsSelf && !ln.selfListen {
			continue
		}
		emit(ctx, ln.ch.Reaction, reaction)
	}
	for _, r := range eventData.Data.GroupReactions {
		reaction := model.NewReaction(uid, r, model.ThreadTypeGroup)
		if reaction.IsSelf && !ln.selfListen {
			continue
		}
		emit(ctx, ln.ch.Reaction, reaction)
	}
}

func (ln *listener) handleOldReactions(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.ReactionEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data", "listener.handleOldReactions", err)
		ln.emitError(ctx, err)
		return
	}

	reactions := make([]model.Reaction, 0, len(eventData.Data.Reactions)+len(eventData.Data.GroupReactions))

	threadType := model.ThreadTypeUser
	selected := eventData.Data.Reactions
	if len(eventData.Data.GroupReactions) > 0 {
		threadType = model.ThreadTypeGroup
		selected = eventData.Data.GroupReactions
	}

	uid := ln.sc.UID()
	for _, r := range selected {
		reactionObject := model.NewReaction(uid, r, threadType)
		reactions = append(reactions, reactionObject)
	}

	emit(ctx, ln.ch.OldReactions, model.NewOldReactions(reactions, threadType))
}

func (ln *listener) handleActions(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.ActionEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleActions", err)
		ln.emitError(ctx, err)
		return
	}

	for _, action := range eventData.Data.Actions {
		switch action.ActionType {
		case "typing":
			// Always triggered by others; no self-check needed			switch action.Action {
			switch action.Action {
			case "typing":
				typingObject := model.NewUserTyping(action.Data.Typing)
				emit(ctx, ln.ch.Typing, model.Typing(typingObject))
			case "gtyping":
				typingObject := model.NewGroupTyping(action.Data.GroupTyping)
				emit(ctx, ln.ch.Typing, model.Typing(typingObject))
			}
		}
	}
}

func (ln *listener) handleControls(ctx context.Context, body BaseWSMessage) {
	eventData, err := decodeEventData[events.ControlEventData](body, ln.cipherKey)
	if err != nil {
		err = errs.WrapZCA("Failed to decode event data:", "listener.handleControls", err)
		ln.emitError(ctx, err)
		return
	}

	for _, ctrl := range eventData.Data.Controls {
		content := ctrl.Content
		switch content.ActionType {
		case "file_done":
			if content.FileID == nil || content.Data.UploadAttachment == nil {
				continue
			}

			fileID, url := *content.FileID, content.Data.UploadAttachment.URL
			uploadObject := model.NewUploadAttachment(strconv.FormatInt(fileID, 10), url)

			uploadCallback, ok := ln.sc.UploadCallback().Get(strconv.FormatInt(fileID, 10))
			if ok {
				go uploadCallback(uploadObject)
				ln.sc.UploadCallback().Delete(strconv.FormatInt(fileID, 10))
			}

			emit(ctx, ln.ch.UploadAttachment, uploadObject)
		case "group":
			if content.Data.GroupEvent == nil {
				continue
			}

			// 31/08/2024
			// for some reason, Zalo send both join and join_reject event when admin approve join requests
			// Zalo itself doesn't seem to handle this properly either, so we gonna ignore the join_reject event
			if content.Action == "join_reject" {
				continue
			}

			groupEvent := model.NewGroupEvent(ln.sc.UID(), content.Action, content.Data.GroupEvent)
			if groupEvent.IsSelf() && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Group, groupEvent)
		case "fr":
			if content.Data.FriendEvent == nil {
				continue
			}

			// 31/08/2024
			// for some reason, Zalo send both join and join_reject event when admin approve join requests
			// Zalo itself doesn't seem to handle this properly either, so we gonna ignore the join_reject event
			if content.Action == "req" {
				continue
			}

			friendEvent := model.NewFriendEvent(ln.sc.UID(), content.Action, content.Data.FriendEvent)
			if friendEvent.IsSelf() && !ln.selfListen {
				continue
			}
			emit(ctx, ln.ch.Friend, friendEvent)
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

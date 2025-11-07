package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Amrakk/zcago/model"
)

func (a *ListenerApp) handleMessage(msg model.Message) {
	timestamp := time.Now().Format("15:04:05")

	var content, displayName string
	switch m := msg.(type) {
	case model.UserMessage:
		if m.Data.Content.String != nil {
			content = *m.Data.Content.String
		} else if m.Data.Content.Attachment != nil {
			content = fmt.Sprintf("[Attachment: %s]", m.Data.Content.Attachment.Title)
		} else {
			content = "[Other content type]"
		}
		displayName = m.Data.DName

	case model.GroupMessage:
		if m.Data.Content.String != nil {
			content = *m.Data.Content.String
		} else if m.Data.Content.Attachment != nil {
			content = fmt.Sprintf("[Attachment: %s]", m.Data.Content.Attachment.Title)
		} else {
			content = "[Other content type]"
		}
		displayName = m.Data.DName

	default:
		content = "[Unknown message type]"
		displayName = "Unknown"
	}

	selfIndicator := ""
	if msg.IsSelf() {
		selfIndicator = " (self)"
	}

	fmt.Printf("📨 [%s] Message%s from %s in %s: %s\n",
		timestamp, selfIndicator, displayName, msg.ThreadID(), content)

	if a.isDebug {
		if data, err := json.MarshalIndent(msg, "", "  "); err == nil {
			fmt.Printf("   Full message data: %s\n", string(data))
		}
	}
}

func (a *ListenerApp) handleOldMessages(oldMessages model.OldMessages) {
	timestamp := time.Now().Format("15:04:05")

	for _, om := range oldMessages.Messages {
		selfIndicator := ""
		if om.IsSelf() {
			selfIndicator = " (self)"
		}

		var content, msgID, cliMsgID string

		switch m := om.(type) {
		case model.UserMessage:
			if m.Data.Content.String != nil {
				content = *m.Data.Content.String
			}
			msgID = m.Data.MsgID
			cliMsgID = m.Data.CliMsgID

		case model.GroupMessage:
			if m.Data.Content.String != nil {
				content = *m.Data.Content.String
			}
			msgID = m.Data.MsgID
			cliMsgID = m.Data.CliMsgID

		default:
			content = "[Unknown old message type]"
		}

		fmt.Printf("🕰️ [%s] Old Message%s in %s - MsgID: %s, CliMsgID: %s: %s\n",
			timestamp, selfIndicator, om.ThreadID(),
			msgID, cliMsgID, content)

	}

	if a.isDebug {
		if data, err := json.MarshalIndent(oldMessages, "", "  "); err == nil {
			fmt.Printf("   Full old messages data: %s\n", string(data))
		}
	}
}

func (a *ListenerApp) handleReaction(reaction model.Reaction) {
	timestamp := time.Now().Format("15:04:05")

	selfIndicator := ""
	if reaction.IsSelf {
		selfIndicator = " (self)"
	}

	threadType := "User"
	if reaction.Type == model.ThreadTypeGroup {
		threadType = "Group"
	}

	fmt.Printf("🙂 [%s] Reaction%s in %s - MsgID: %s, CliMsgID: %s, RType: %d, RIcon: %s\n",
		timestamp, selfIndicator, threadType,
		reaction.Data.MsgID, reaction.Data.CliMsgID,
		reaction.Data.Content.RType, reaction.Data.Content.RIcon)

	if a.isDebug {
		if data, err := json.MarshalIndent(reaction, "", "  "); err == nil {
			fmt.Printf("   Full reaction data: %s\n", string(data))
		}
	}
}

func (a *ListenerApp) handleOldReactions(old model.OldReactions) {
	timestamp := time.Now().Format("15:04:05")

	for _, r := range old.Reactions {
		self := ""
		if r.IsSelf {
			self = " (self)"
		}

		thread := "User"
		if r.Type == model.ThreadTypeGroup {
			thread = "Group"
		}

		fmt.Printf("🕰️ [%s] Old Reaction%s in %s - MsgID: %s, CliMsgID: %s, RType: %d, RIcon: %s\n",
			timestamp, self, thread,
			r.Data.MsgID, r.Data.CliMsgID,
			r.Data.Content.RType, r.Data.Content.RIcon)
	}

	if a.isDebug {
		if data, err := json.MarshalIndent(old, "", "  "); err == nil {
			fmt.Printf("   Full old reactions data: %s\n", string(data))
		}
	}
}

func (a *ListenerApp) handleTyping(typing model.Typing) {
	timestamp := time.Now().Format("15:04:05")

	var content, displayName string
	switch m := typing.(type) {
	case model.UserTyping:
		content = fmt.Sprintf("User Typing - UID: %s, TS: %s, IsPC: %d", m.Data.UID, m.Data.TS, m.Data.IsPC)
		displayName = m.Data.UID
	case model.GroupTyping:
		content = fmt.Sprintf("Group Typing - GID: %s, TS: %s, IsPC: %d", m.Data.GID, m.Data.TS, m.Data.IsPC)
		displayName = m.Data.GID
	default:
		content = "[Unknown typing type]"
		displayName = "Unknown"
	}

	selfIndicator := ""
	if typing.IsSelf() {
		selfIndicator = " (self)"
	}

	fmt.Printf("⌨️ [%s] Typing%s from %s: %s\n",
		timestamp, selfIndicator, displayName, content)

	if a.isDebug {
		if data, err := json.MarshalIndent(typing, "", "  "); err == nil {
			fmt.Printf("   Full typing data: %s\n", string(data))
		}
	}
}

func (a *ListenerApp) handleDeliveredMessages(deliveredMsgs []model.DeliveredMessage) {
	timestamp := time.Now().Format("15:04:05")

	for _, delivered := range deliveredMsgs {
		selfIndicator := ""
		if delivered.IsSelf() {
			selfIndicator = " (self)"
		}

		var summary string
		switch m := delivered.(type) {
		case model.UserDeliveredMessage:
			summary = fmt.Sprintf("User Delivered — UIDs: %v, MsgID: %v",
				m.Data.DeliveredUIDs, m.Data.MsgID)

		case model.GroupDeliveredMessage:
			summary = fmt.Sprintf("Group Delivered — GID: %s, UIDs: %v, MsgID: %v",
				m.Data.GroupID, m.Data.DeliveredUIDs, m.Data.MsgID)

		default:
			summary = "[Unknown delivered message type]"
		}

		fmt.Printf("✅ [%s] Delivered Message%s in %s: %s\n",
			timestamp, selfIndicator, delivered.ThreadID(), summary)

		if a.isDebug {
			if data, err := json.MarshalIndent(delivered, "", "  "); err == nil {
				fmt.Printf("   Full delivered message data: %s\n", string(data))
			}
		}
	}
}

func (a *ListenerApp) handleSeenMessages(seenMsgs []model.SeenMessage) {
	timestamp := time.Now().Format("15:04:05")

	for _, seen := range seenMsgs {
		selfIndicator := ""
		if seen.IsSelf() {
			selfIndicator = " (self)"
		}

		var summary string
		switch m := seen.(type) {
		case model.UserSeenMessage:
			summary = fmt.Sprintf("User Seen — IDTo: %s, MsgID: %s", m.Data.IDTo, m.Data.MsgID)

		case model.GroupSeenMessage:
			summary = fmt.Sprintf("Group Seen — GID: %s, SeenUIDs: %v, MsgID: %s",
				m.Data.GroupID, m.Data.SeenUIDs, m.Data.MsgID)

		default:
			summary = "[Unknown seen message type]"
		}

		fmt.Printf("👁️ [%s] Seen Message%s in %s: %s\n",
			timestamp, selfIndicator, seen.ThreadID(), summary)

		if a.isDebug {
			if data, err := json.MarshalIndent(seen, "", "  "); err == nil {
				fmt.Printf("   Full seen message data: %s\n", string(data))
			}
		}
	}
}

func (a *ListenerApp) handleUndo(undo model.Undo) {
	timestamp := time.Now().Format("15:04:05")

	selfIndicator := ""
	if undo.IsSelf {
		selfIndicator = " (self)"
	}
	groupIndicator := ""
	if undo.IsGroup {
		groupIndicator = " [Group]"
	}

	fmt.Printf("↩️ [%s] Undo%s%s in %s - GlobalMsgID: %v, CliMsgID: %v\n",
		timestamp, selfIndicator, groupIndicator, undo.ThreadID,
		undo.Data.Content.GlobalMsgID, undo.Data.Content.CliMsgID)

	if a.isDebug {
		if data, err := json.MarshalIndent(undo, "", "  "); err == nil {
			fmt.Printf("   Full undo data: %s\n", string(data))
		}
	}
}

func (a *ListenerApp) handleGroup(ev model.GroupEvent) {
	timestamp := time.Now().Format("15:04:05")

	selfIndicator := ""
	if ev.IsSelf() {
		selfIndicator = " (self)"
	}

	eventType := ev.Type()
	action := ev.Action()
	threadID := ev.ThreadID()
	data := ev.Data()

	category := "Unknown"
	switch eventType {
	case model.GroupEventTypeJoinRequest:
		category = "Join Request"
	case model.GroupEventTypeJoin:
		category = "Member Joined"
	case model.GroupEventTypeLeave, model.GroupEventTypeRemoveMember, model.GroupEventTypeBlockMember:
		category = "Member Left/Removed"
	case model.GroupEventTypeNewPinTopic, model.GroupEventTypeUpdatePinTopic, model.GroupEventTypeUnpinTopic:
		category = "Pinned Topic"
	case model.GroupEventTypeReorderPinTopic:
		category = "Reorder Pinned Topics"
	case model.GroupEventTypeUpdateBoard, model.GroupEventTypeRemoveBoard:
		category = "Board Update"
	case model.GroupEventTypeAcceptRemind, model.GroupEventTypeRejectRemind:
		category = "Reminder Response"
	case model.GroupEventTypeRemindTopic:
		category = "Reminder Topic"
	case model.GroupEventTypeUpdateAvatar:
		category = "Avatar Update"
	}

	detail := ""
	switch v := data.(type) {
	case model.TGroupEventJoinRequest:
		detail = fmt.Sprintf("Pending: %d, UIDs: %v", v.TotalPending, v.UIDs)

	case model.TGroupEventPinTopic:
		detail = fmt.Sprintf("Actor: %s, TopicID: %s, TopicType: %d, BoardVersion: %d (old: %d)",
			v.ActorID, v.Topic.ID, v.Topic.Type, v.BoardVersion, v.OldBoardVersion)

	case model.TGroupEventReorderPinTopic:
		detail = fmt.Sprintf("Actor: %s, Topics: %d, BoardVersion: %d (old: %d)",
			v.ActorID, len(v.Topics), v.BoardVersion, v.OldBoardVersion)

	case model.TGroupEventBoard:
		detail = fmt.Sprintf("GroupName: %s, SourceID: %s", v.GroupName, v.SourceID)

	case model.TGroupEventRemindRespond:
		detail = fmt.Sprintf("TopicID: %s, Members: %v", v.TopicID, v.UpdateMembers)

	case model.TGroupEventRemindTopic:
		detail = fmt.Sprintf("Creator: %s, Msg: %q, Color: %s, Emoji: %s",
			v.CreatorID, v.Msg, v.Color, v.Emoji)

	case model.TGroupEventBase:
		detail = fmt.Sprintf("GroupName: %s, SubType: %d", v.GroupName, v.SubType)
	}

	if detail != "" {
		fmt.Printf("👥 [%s] Group Event%s - Category: %s, Type: %s, ThreadID: %s, Action: %s | %s\n",
			timestamp, selfIndicator, category, eventType, threadID, action, detail)
	} else {
		fmt.Printf("👥 [%s] Group Event%s - Category: %s, Type: %s, ThreadID: %s, Action: %s\n",
			timestamp, selfIndicator, category, eventType, threadID, action)
	}

	if a.isDebug {
		if dataBytes, err := json.MarshalIndent(ev.Data(), "", "  "); err == nil {
			fmt.Printf("   Full group event data: %s\n", string(dataBytes))
		}
	}
}

func (a *ListenerApp) handleFriend(ev model.FriendEvent) {
	timestamp := time.Now().Format("15:04:05")

	selfIndicator := ""
	if ev.IsSelf() {
		selfIndicator = " (self)"
	}

	eventType := ev.Type()
	action := ev.Action()
	threadID := ev.ThreadID()
	data := ev.Data()

	category := "Unknown"
	switch eventType {
	case model.FriendEventTypeAdd:
		category = "Friend Added"
	case model.FriendEventTypeRemove:
		category = "Friend Removed"
	case model.FriendEventTypeRequest:
		category = "Friend Request"
	case model.FriendEventTypeUndoRequest:
		category = "Friend Request Undone"
	case model.FriendEventTypeSeenFriendRequest:
		category = "Friend Request Seen"
	case model.FriendEventTypeRejectRequest:
		category = "Friend Request Rejected"
	case model.FriendEventTypeBlock:
		category = "Friend Blocked"
	case model.FriendEventTypeUnblock:
		category = "Friend Unblocked"
	case model.FriendEventTypeBlockCall:
		category = "Friend Call Blocked"
	case model.FriendEventTypeUnblockCall:
		category = "Friend Call Unblocked"
	case model.FriendEventTypePinCreate:
		category = "Friend Pinned"
	case model.FriendEventTypePinUnpin:
		category = "Friend Unpinned"
	}

	detail := ""
	switch v := data.(type) {
	case model.TFriendEventBase:
		detail = fmt.Sprintf("FriendID: %s", v.FriendID())

	case model.TFriendEventRequest:
		detail = fmt.Sprintf("From: %s → To: %s, Src: %d, Message: %q",
			v.FromUID, v.ToUID, v.Src, v.Message)

	case model.TFriendEventRejectUndo:
		detail = fmt.Sprintf("From: %s → To: %s", v.FromUID, v.ToUID)

	case model.TFriendEventSeenRequest:
		detail = fmt.Sprintf("Seen UIDs: %v", []string(v))

	case model.TFriendEventPinCreate:
		detail = fmt.Sprintf("Actor: %s, TopicID: %s, ConvID: %s, Version: %d (old: %d)",
			v.ActorID, v.Topic.ID, v.ConversationID, v.Version, v.OldVersion)

	case model.TFriendEventPinUnpin:
		detail = fmt.Sprintf("Actor: %s, TopicID: %s, ConvID: %s, Version: %d (old: %d)",
			v.ActorID, v.Topic.TopicID, v.ConversationID, v.Version, v.OldVersion)
	}

	if detail != "" {
		fmt.Printf("👤 [%s] Friend Event%s - Category: %s, Type: %s, ThreadID: %s, Action: %s | %s\n",
			timestamp, selfIndicator, category, eventType, threadID, action, detail)
	} else {
		fmt.Printf("👤 [%s] Friend Event%s - Category: %s, Type: %s, ThreadID: %s, Action: %s\n",
			timestamp, selfIndicator, category, eventType, threadID, action)
	}

	if a.isDebug {
		if dataBytes, err := json.MarshalIndent(ev.Data(), "", "  "); err == nil {
			fmt.Printf("   Full friend event data: %s\n", string(dataBytes))
		}
	}
}

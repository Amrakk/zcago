package model

import "slices"

type UserSeenMessage struct {
	Type     ThreadType
	Data     TUserSeenMessage
	ThreadID string
	IsSelf   bool
}

func NewUserSeenMessage(data TUserSeenMessage) UserSeenMessage {
	return UserSeenMessage{
		Type:     ThreadTypeUser,
		Data:     data,
		ThreadID: data.IDTo,
		IsSelf:   false,
	}
}

type GroupSeenMessage struct {
	Type     ThreadType
	Data     TGroupSeenMessage
	ThreadID string
	IsSelf   bool
}

func NewGroupSeenMessage(uid string, data TGroupSeenMessage) GroupSeenMessage {
	return GroupSeenMessage{
		Type:     ThreadTypeGroup,
		Data:     data,
		ThreadID: data.GroupID,
		IsSelf:   slices.Contains(data.SeenUIDs, uid),
	}
}

type TUserSeenMessage struct {
	IDTo      string `json:"idTo"`
	MsgID     string `json:"msgId"`
	RealMsgID string `json:"realMsgId"`
}

type TGroupSeenMessage struct {
	MsgID    string   `json:"msgId"`
	GroupID  string   `json:"groupId"`
	SeenUIDs []string `json:"seenUids"`
}

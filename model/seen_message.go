package model

import "slices"

type SeenMessage interface {
	Type() ThreadType
	ThreadID() string
	IsSelf() bool
}

type UserSeenMessage struct {
	typ      ThreadType
	Data     TUserSeenMessage
	threadID string
	isSelf   bool
}

func NewUserSeenMessage(data TUserSeenMessage) UserSeenMessage {
	return UserSeenMessage{
		typ:      ThreadTypeUser,
		Data:     data,
		threadID: data.IDTo,
		isSelf:   false,
	}
}

func (usm UserSeenMessage) Type() ThreadType { return usm.typ }
func (usm UserSeenMessage) ThreadID() string { return usm.threadID }
func (usm UserSeenMessage) IsSelf() bool     { return usm.isSelf }

type GroupSeenMessage struct {
	typ      ThreadType
	Data     TGroupSeenMessage
	threadID string
	isSelf   bool
}

func NewGroupSeenMessage(uid string, data TGroupSeenMessage) GroupSeenMessage {
	return GroupSeenMessage{
		typ:      ThreadTypeGroup,
		Data:     data,
		threadID: data.GroupID,
		isSelf:   slices.Contains(data.SeenUIDs, uid),
	}
}

func (gsm GroupSeenMessage) Type() ThreadType { return gsm.typ }
func (gsm GroupSeenMessage) ThreadID() string { return gsm.threadID }
func (gsm GroupSeenMessage) IsSelf() bool     { return gsm.isSelf }

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

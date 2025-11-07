package model

import "slices"

type DeliveredMessage interface {
	Type() ThreadType
	ThreadID() string
	IsSelf() bool
}

type UserDeliveredMessage struct {
	typ      ThreadType
	Data     TDeliveredMessage
	threadID string
	isSelf   bool
}

func NewUserDeliveredMessage(data TDeliveredMessage) UserDeliveredMessage {
	return UserDeliveredMessage{
		typ:      ThreadTypeUser,
		Data:     data,
		threadID: data.DeliveredUIDs[0],
		isSelf:   false,
	}
}

func (udm UserDeliveredMessage) Type() ThreadType { return udm.typ }
func (udm UserDeliveredMessage) ThreadID() string { return udm.threadID }
func (udm UserDeliveredMessage) IsSelf() bool     { return udm.isSelf }

type GroupDeliveredMessage struct {
	typ      ThreadType
	Data     TGroupDeliveredMessage
	threadID string
	isSelf   bool
}

func NewGroupDeliveredMessage(uid string, data TGroupDeliveredMessage) GroupDeliveredMessage {
	return GroupDeliveredMessage{
		typ:      ThreadTypeGroup,
		Data:     data,
		threadID: data.GroupID,
		isSelf:   slices.Contains(data.DeliveredUIDs, uid),
	}
}

func (gdm GroupDeliveredMessage) Type() ThreadType { return gdm.typ }
func (gdm GroupDeliveredMessage) ThreadID() string { return gdm.threadID }
func (gdm GroupDeliveredMessage) IsSelf() bool     { return gdm.isSelf }

type TDeliveredMessage struct {
	MsgID         string   `json:"msgId"`
	Seen          int      `json:"seen"`
	DeliveredUIDs []string `json:"deliveredUids"`
	SeenUIDs      []string `json:"seenUids"`
	RealMsgID     string   `json:"realMsgId"`
	MSTs          int      `json:"mSTs"`
}

type TGroupDeliveredMessage struct {
	TDeliveredMessage
	GroupID string `json:"groupId"`
}

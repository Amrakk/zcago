package model

import "slices"

type UserDeliveredMessage struct {
	Type     ThreadType
	Data     TDeliveredMessage
	ThreadID string
	IsSelf   bool
}

func NewUserDeliveredMessage(data TDeliveredMessage) UserDeliveredMessage {
	return UserDeliveredMessage{
		Type:     ThreadTypeUser,
		Data:     data,
		ThreadID: data.DeliveredUIDs[0],
		IsSelf:   false,
	}
}

type GroupDeliveredMessage struct {
	Type     ThreadType
	Data     TGroupDeliveredMessage
	ThreadID string
	IsSelf   bool
}

func NewGroupDeliveredMessage(uid string, data TGroupDeliveredMessage) GroupDeliveredMessage {
	return GroupDeliveredMessage{
		Type:     ThreadTypeGroup,
		Data:     data,
		ThreadID: data.GroupID,
		IsSelf:   slices.Contains(data.DeliveredUIDs, uid),
	}
}

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

package model

type UserTyping struct {
	Type     ThreadType
	Data     TTyping
	ThreadID string
	IsSelf   bool
}

func NewUserTyping(data TTyping) UserTyping {
	return UserTyping{
		Type:     ThreadTypeUser,
		Data:     data,
		ThreadID: data.UID,
		IsSelf:   false,
	}
}

type GroupTyping struct {
	Type     ThreadType
	Data     TGroupTyping
	ThreadID string
	IsSelf   bool
}

func NewGroupTyping(data TGroupTyping) GroupTyping {
	return GroupTyping{
		Type:     ThreadTypeGroup,
		Data:     data,
		ThreadID: data.GID,
		IsSelf:   false,
	}
}

type TTyping struct {
	UID  string `json:"uid"`
	TS   string `json:"ts"`
	IsPC int    `json:"isPC"` // 0 | 1
}

type TGroupTyping struct {
	TTyping
	GID string `json:"gid"`
}

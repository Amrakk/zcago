package model

type Typing interface {
	Type() ThreadType
	ThreadID() string
	IsSelf() bool
}

type UserTyping struct {
	typ      ThreadType
	Data     TTyping
	threadID string
	isSelf   bool
}

func NewUserTyping(data TTyping) UserTyping {
	return UserTyping{
		typ:      ThreadTypeUser,
		Data:     data,
		threadID: data.UID,
		isSelf:   false,
	}
}

func (t UserTyping) Type() ThreadType { return t.typ }
func (t UserTyping) ThreadID() string { return t.threadID }
func (t UserTyping) IsSelf() bool     { return t.isSelf }

type GroupTyping struct {
	typ      ThreadType
	Data     TGroupTyping
	threadID string
	isSelf   bool
}

func NewGroupTyping(data TGroupTyping) GroupTyping {
	return GroupTyping{
		typ:      ThreadTypeGroup,
		Data:     data,
		threadID: data.GID,
		isSelf:   false,
	}
}

func (t GroupTyping) Type() ThreadType { return t.typ }
func (t GroupTyping) ThreadID() string { return t.threadID }
func (t GroupTyping) IsSelf() bool     { return t.isSelf }

type TTyping struct {
	UID  string `json:"uid"`
	TS   string `json:"ts"`
	IsPC uint8  `json:"isPC"`
}

type TGroupTyping struct {
	TTyping
	GID string `json:"gid"`
}

package model

import "encoding/json"

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
	IsPC bool   `json:"isPC"`
}

func (t *TTyping) UnmarshalJSON(data []byte) error {
	type Alias TTyping
	aux := &struct {
		IsPC int `json:"isPC"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.IsPC = aux.IsPC != 0
	return nil
}

type TGroupTyping struct {
	TTyping
	GID string `json:"gid"`
}

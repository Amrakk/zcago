package model

import (
	"encoding/json"

	"github.com/Amrakk/zcago/errs"
)

type ReminderRepeatMode int

const (
	RepeatNone ReminderRepeatMode = iota
	RepeatDaily
	RepeatWeekly
	RepeatMonthly
)

type ReminderUser struct {
	ReminderID string             `json:"reminderId"`
	CreatorUID string             `json:"creatorUid"`
	ToUID      string             `json:"toUid"`
	Emoji      string             `json:"emoji"`
	Color      int                `json:"color"`
	Type       int                `json:"type"`
	Repeat     ReminderRepeatMode `json:"repeat"`
	Params     ReminderParams     `json:"params"`
	StartTime  int64              `json:"startTime"`
	EndTime    int64              `json:"endTime"`
	CreateTime int64              `json:"createTime"`
	EditTime   int64              `json:"editTime"`
}

type ReminderGroup struct {
	ID          string             `json:"id"`
	GroupID     string             `json:"groupId"`
	EditorID    string             `json:"editorId"`
	CreatorID   string             `json:"creatorId"`
	Emoji       string             `json:"emoji"`
	Color       int                `json:"color"`
	Type        int                `json:"type"`
	EventType   int                `json:"eventType"`
	Params      ReminderParams     `json:"-"` // Handle this manually
	ResponseMem ResponseMembers    `json:"responseMem"`
	RepeatData  []any              `json:"repeatData"`
	RepeatInfo  *RepeatInfo        `json:"repeatInfo,omitempty"`
	Repeat      ReminderRepeatMode `json:"repeat"`
	Duration    int64              `json:"duration"`
	StartTime   int64              `json:"startTime"`
	CreateTime  int64              `json:"createTime"`
	EditTime    int64              `json:"editTime"`
}

func (rg *ReminderGroup) UnmarshalJSON(data []byte) error {
	type alias ReminderGroup
	aux := &struct {
		Params json.RawMessage `json:"params"`
		*alias
	}{
		alias: (*alias)(rg),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Params) > 0 && string(aux.Params) != "null" {
		var paramsStr string
		if err := json.Unmarshal(aux.Params, &paramsStr); err == nil {
			var params ReminderParams
			if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
				return err
			}
			rg.Params = params
		} else {
			var params ReminderParams
			if err := json.Unmarshal(aux.Params, &params); err != nil {
				return err
			}
			rg.Params = params
		}
	}

	return nil
}

type ReminderParams struct {
	Title    string `json:"title"`
	SetTitle bool   `json:"setTitle,omitempty"`
}

type ResponseMembers struct {
	RejectMember int `json:"rejectMember"`
	MyResp       int `json:"myResp"`
	AcceptMember int `json:"acceptMember"`
}

type RepeatInfo struct {
	ListTS []any `json:"list_ts"`
}

var ErrAmbiguousReminder = errs.NewZCA("ambiguous empty array (cannot determine user/group)", "model.ReminderResponse.UnmarshalJSON")

type ReminderResponse[TUser, TGroup any] struct {
	ThreadType ThreadType
	usr        *TUser
	grp        *TGroup
}

func (r *ReminderResponse[TUser, TGroup]) User() *TUser {
	if r.ThreadType == ThreadTypeUser {
		return r.usr
	}
	return nil
}

func (r *ReminderResponse[TUser, TGroup]) Group() *TGroup {
	if r.ThreadType == ThreadTypeGroup {
		return r.grp
	}
	return nil
}

func (r *ReminderResponse[TUser, TGroup]) SetUser(user *TUser) {
	r.ThreadType = ThreadTypeUser
	r.usr = user
}

func (r *ReminderResponse[TUser, TGroup]) SetGroup(group *TGroup) {
	r.ThreadType = ThreadTypeGroup
	r.grp = group
}

func (r *ReminderResponse[TUser, TGroup]) UnmarshalJSON(data []byte) error {
	*r = ReminderResponse[TUser, TGroup]{}

	data = r.unwrapStringJSON(data)

	if data[0] == '[' {
		firstItem, err := r.extractFirstArrayItem(data)
		if err != nil {
			return err
		}
		data = firstItem
	}

	return r.unmarshalByProbe(data)
}

func (r *ReminderResponse[TUser, TGroup]) unwrapStringJSON(data []byte) []byte {
	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return data
		}
		return []byte(raw)
	}
	return data
}

func (r *ReminderResponse[TUser, TGroup]) extractFirstArrayItem(data []byte) ([]byte, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, ErrAmbiguousReminder
	}

	return items[0], nil
}

func (r *ReminderResponse[TUser, TGroup]) unmarshalByProbe(data []byte) error {
	var probe struct {
		ID    string `json:"id"`
		ToUID string `json:"toUid"`
	}

	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}

	if probe.ToUID != "" && probe.ID != "" {
		return ErrAmbiguousReminder
	}

	switch {
	case probe.ToUID != "":
		return r.unmarshalAsUser(data)
	case probe.ID != "":
		return r.unmarshalAsGroup(data)
	default:
		return ErrAmbiguousReminder
	}
}

func (r *ReminderResponse[TUser, TGroup]) unmarshalAsUser(data []byte) error {
	var user TUser
	if err := json.Unmarshal(data, &user); err != nil {
		return err
	}
	r.SetUser(&user)
	return nil
}

func (r *ReminderResponse[TUser, TGroup]) unmarshalAsGroup(data []byte) error {
	var group TGroup
	if err := json.Unmarshal(data, &group); err != nil {
		return err
	}
	r.SetGroup(&group)
	return nil
}

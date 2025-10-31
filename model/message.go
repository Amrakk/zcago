package model

import (
	"encoding/json"
	"strconv"

	"github.com/Amrakk/zcago/errs"
)

type Urgency int

const (
	UrgDefault Urgency = iota
	UrgImportant
	UrgUrgent
)

type UserMessage struct {
	Type     ThreadType
	Data     TMessage
	ThreadID string
	IsSelf   bool
}

func NewUserMessage(uid string, data TMessage) UserMessage {
	msg := UserMessage{
		Type:     ThreadTypeUser,
		Data:     data,
		ThreadID: data.UIDFrom,
		IsSelf:   data.UIDFrom == "0",
	}

	if data.UIDFrom == "0" {
		msg.ThreadID = data.IDTo
	}
	if data.IDTo == "0" {
		msg.Data.IDTo = uid
	}
	if data.UIDFrom == "0" {
		msg.Data.UIDFrom = uid
	}

	return msg
}

type GroupMessage struct {
	Type     ThreadType
	Data     TGroupMessage
	ThreadID string
	IsSelf   bool
}

func NewGroupMessage(uid string, data TGroupMessage) *GroupMessage {
	g := &GroupMessage{
		Type:     ThreadTypeGroup,
		Data:     data,
		ThreadID: data.IDTo,
		IsSelf:   data.UIDFrom == "0",
	}

	if data.UIDFrom == "0" {
		g.Data.UIDFrom = uid
	}

	return g
}

type TMessage struct {
	ActionID          string       `json:"actionId"`
	MsgID             string       `json:"msgId"`
	CliMsgID          string       `json:"cliMsgId"`
	MsgType           string       `json:"msgType"`
	UIDFrom           string       `json:"uidFrom"`
	IDTo              string       `json:"idTo"`
	DName             string       `json:"dName"`
	TS                string       `json:"ts"`
	Status            int          `json:"status"`
	Content           Content      `json:"content"`
	Notify            string       `json:"notify"`
	TTL               int          `json:"ttl"`
	UserID            string       `json:"userId"`
	UIN               string       `json:"uin"`
	TopOut            string       `json:"topOut"`
	TopOutTimeOut     string       `json:"topOutTimeOut"`
	TopOutImprTimeOut string       `json:"topOutImprTimeOut"`
	PropertyExt       *PropertyExt `json:"propertyExt,omitempty"`
	ParamsExt         ParamsExt    `json:"paramsExt"`
	CMD               int          `json:"cmd"`
	ST                int          `json:"st"`
	AT                int          `json:"at"`
	RealMsgID         string       `json:"realMsgId"`
	Quote             *TQuote      `json:"quote,omitempty"`
}

type TGroupMessage struct {
	TMessage
	Mentions []*TMention `json:"mentions,omitempty"`
}

type PropertyExt struct {
	Color   int    `json:"color"`
	Size    int    `json:"size"`
	Type    int    `json:"type"`
	SubType int    `json:"subType"`
	Ext     string `json:"ext"`
}

type ParamsExt struct {
	CountUnread  int `json:"countUnread"`
	ContainType  int `json:"containType"`
	PlatformType int `json:"platformType"`
}

type TQuote struct {
	OwnerID     string `json:"ownerId"`
	CliMsgID    int64  `json:"cliMsgId"`
	GlobalMsgID int64  `json:"globalMsgId"`
	CliMsgType  int    `json:"cliMsgType"`
	Timestamp   int64  `json:"ts"`
	Msg         string `json:"msg"`
	Attach      string `json:"attach"`
	FromD       string `json:"fromD"`
	TTL         uint   `json:"ttl"`
}

func (tq *TQuote) UnmarshalJSON(data []byte) error {
	type alias TQuote
	aux := &struct {
		OwnerID int `json:"ownerId"`
		*alias
	}{
		alias: (*alias)(tq),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	tq.OwnerID = strconv.Itoa(aux.OwnerID)
	return nil
}

type TAttachmentContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Href        string `json:"href"`
	Thumb       string `json:"thumb"`
	ChildNumber int    `json:"childnumber"`
	Action      string `json:"action"`
	Params      string `json:"params"`
	Type        string `json:"type"`
}

type TOtherContent map[string]any

type MentionType int

const (
	MentionEach MentionType = iota
	MentionAll
)

type TMention struct {
	UID  string      `json:"uid"` // User ID being mentioned, or "-1" for mention all
	Pos  int         `json:"pos"` // Mention position
	Len  int         `json:"len"` // Mention length
	Type MentionType `json:"type"`
}

func (m *TMention) IsValid() bool {
	if m.Type == 1 && m.UID == "-1" {
		return true
	}
	if m.Type == 0 && m.UID != "" && m.UID != "-1" && m.Len > 0 {
		return true
	}
	return false
}

type Content struct {
	String     *string
	Attachment *TAttachmentContent
	Other      TOtherContent
}

func (c *Content) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.String = &s
		return nil
	}

	var attach TAttachmentContent
	if err := json.Unmarshal(data, &attach); err == nil && attach.Title != "" {
		c.Attachment = &attach
		return nil
	}

	var other TOtherContent
	if err := json.Unmarshal(data, &other); err == nil {
		c.Other = other
		return nil
	}

	return errs.NewZCA("Content: data did not match any known content type", "Content.UnmarshalJSON")
}

func (c Content) MarshalJSON() ([]byte, error) {
	if c.String != nil {
		return json.Marshal(c.String)
	}
	if c.Attachment != nil {
		return json.Marshal(c.Attachment)
	}
	if c.Other != nil {
		return json.Marshal(c.Other)
	}
	return []byte("null"), nil
}

type OutboundMessage struct {
	MsgID    string
	CliMsgID string
	UIDFrom  string
	IDTo     string
	MsgType  string
	ST       int
	AT       int
	CMD      int
	TS       int64
}

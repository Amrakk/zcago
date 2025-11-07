package model

import "encoding/json"

type FriendEventType string

const (
	FriendEventTypeAdd    FriendEventType = "add"
	FriendEventTypeRemove FriendEventType = "remove"

	FriendEventTypeRequest           FriendEventType = "req_v2"
	FriendEventTypeUndoRequest       FriendEventType = "undo_req"
	FriendEventTypeRejectRequest     FriendEventType = "reject"
	FriendEventTypeSeenFriendRequest FriendEventType = "seen_fr_req"

	FriendEventTypeBlock       FriendEventType = "block"
	FriendEventTypeUnblock     FriendEventType = "unblock"
	FriendEventTypeBlockCall   FriendEventType = "block_call"
	FriendEventTypeUnblockCall FriendEventType = "unblock_call"

	FriendEventTypePinUnpin  FriendEventType = "pin_unpin"
	FriendEventTypePinCreate FriendEventType = "pin_create"

	FriendEventTypeUnknown FriendEventType = "unknown"
)

func ParseFriendEventType(s string) FriendEventType {
	switch s {
	case "add":
		return FriendEventTypeAdd
	case "remove":
		return FriendEventTypeRemove
	case "req_v2":
		return FriendEventTypeRequest
	case "undo_req":
		return FriendEventTypeUndoRequest
	case "reject":
		return FriendEventTypeRejectRequest
	case "seen_fr_req":
		return FriendEventTypeSeenFriendRequest
	case "block":
		return FriendEventTypeBlock
	case "unblock":
		return FriendEventTypeUnblock
	case "block_call":
		return FriendEventTypeBlockCall
	case "unblock_call":
		return FriendEventTypeUnblockCall
	case "pin_unpin":
		return FriendEventTypePinUnpin
	case "pin_create":
		return FriendEventTypePinCreate
	default:
		return FriendEventTypeUnknown
	}
}

type FriendEvent interface {
	IsSelf() bool
	Data() TFriendEvent
	Type() FriendEventType
	Action() string
	ThreadID() string
}

func NewFriendEvent(uid, action string, data TFriendEvent) FriendEvent {
	frEvent := ParseFriendEventType(action)
	threadID := data.FriendID()
	isSelf := false

	switch frEvent {
	case FriendEventTypeAdd, FriendEventTypeRemove:
		isSelf = false
	case FriendEventTypeBlock, FriendEventTypeUnblock, FriendEventTypeBlockCall, FriendEventTypeUnblockCall:
		isSelf = true
	case FriendEventTypeRejectRequest, FriendEventTypeUndoRequest:
		isSelf = data.(TFriendEventRejectUndo).FromUID == uid
	case FriendEventTypeRequest:
		isSelf = data.(TFriendEventRequest).FromUID == uid
	case FriendEventTypeSeenFriendRequest:
		isSelf = true
		threadID = uid
	case FriendEventTypePinCreate:
		isSelf = data.(TFriendEventPinCreate).ActorID == uid
	case FriendEventTypePinUnpin:
		isSelf = data.(TFriendEventPinUnpin).ActorID == uid
	default:
		isSelf = false
	}

	return friendEvent{
		typ:      frEvent,
		data:     data,
		act:      action,
		threadID: threadID,
		isSelf:   isSelf,
	}
}

type friendEvent struct {
	typ      FriendEventType
	data     TFriendEvent
	act      string
	threadID string
	isSelf   bool
}

func (e friendEvent) Type() FriendEventType { return e.typ }
func (e friendEvent) Data() TFriendEvent    { return e.data }
func (e friendEvent) Action() string        { return e.act }
func (e friendEvent) ThreadID() string      { return e.threadID }
func (e friendEvent) IsSelf() bool          { return e.isSelf }

type TFriendEvent interface {
	EventType() EventType
	FriendID() string
}

type TFriendEventBase string

func (e TFriendEventBase) EventType() EventType { return EventTypeFriend }
func (e TFriendEventBase) FriendID() string     { return string(e) }

type TFriendEventRejectUndo struct {
	ToUID   string `json:"toUid"`
	FromUID string `json:"fromUid"`
}

func (e TFriendEventRejectUndo) EventType() EventType { return EventTypeFriend }
func (e TFriendEventRejectUndo) FriendID() string     { return e.ToUID }

type TFriendEventRequest struct {
	ToUID   string `json:"toUid"`
	FromUID string `json:"fromUid"`
	Src     int    `json:"src"`
	Message string `json:"message"`
}

func (e TFriendEventRequest) EventType() EventType { return EventTypeFriend }
func (e TFriendEventRequest) FriendID() string     { return e.ToUID }

type TFriendEventSeenRequest []string

func (e TFriendEventSeenRequest) EventType() EventType { return EventTypeFriend }
func (e TFriendEventSeenRequest) FriendID() string     { return "" }

type TFriendEventPinUnpin struct {
	Topic          TFriendEventPinTopic `json:"topic"`
	ActorID        string               `json:"actorId"`
	OldVersion     int                  `json:"oldVersion"`
	Version        int                  `json:"version"`
	ConversationID string               `json:"conversationId"`
}

func (e TFriendEventPinUnpin) EventType() EventType { return EventTypeFriend }
func (e TFriendEventPinUnpin) FriendID() string     { return e.ConversationID }

type TFriendEventPinCreate struct {
	OldTopic       *TFriendEventPinTopic      `json:"oldTopic,omitempty"`
	Topic          TFriendEventPinCreateTopic `json:"topic"`
	ActorID        string                     `json:"actorId"`
	OldVersion     int                        `json:"oldVersion"`
	Version        int                        `json:"version"`
	ConversationID string                     `json:"conversationId"`
}

func (e TFriendEventPinCreate) EventType() EventType { return EventTypeFriend }
func (e TFriendEventPinCreate) FriendID() string     { return e.ConversationID }

type TFriendEventPinTopic struct {
	TopicID   string `json:"topicId"`
	TopicType int    `json:"topicType"`
}

type TFriendEventPinCreateTopic struct {
	Type       int                              `json:"type"`
	Color      int                              `json:"color"`
	Emoji      string                           `json:"emoji"`
	StartTime  int64                            `json:"startTime"`
	Duration   int64                            `json:"duration"`
	Params     TFriendEventPinCreateTopicParams `json:"params"`
	ID         string                           `json:"id"`
	CreatorID  string                           `json:"creatorId"`
	CreateTime int64                            `json:"createTime"`
	EditorID   string                           `json:"editorId"`
	EditTime   int64                            `json:"editTime"`
	Repeat     int                              `json:"repeat"`
	Action     int                              `json:"action"`
}

type TFriendEventPinCreateTopicParams struct {
	SenderUID   string `json:"senderUid"`
	SenderName  string `json:"senderName"`
	ClientMsgID string `json:"client_msg_id"`
	GlobalMsgID string `json:"global_msg_id"`
	MsgType     int    `json:"msg_type"`
	Title       string `json:"title"`
}

func (p *TFriendEventPinCreateTopicParams) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		data = []byte(s)
	}

	type alias TFriendEventPinCreateTopicParams
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*p = TFriendEventPinCreateTopicParams(tmp)
	return nil
}

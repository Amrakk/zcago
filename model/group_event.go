package model

import (
	"slices"
)

type GroupEventType string

const (
	GroupEventTypeJoinRequest  GroupEventType = "join_request"
	GroupEventTypeJoin         GroupEventType = "join"
	GroupEventTypeLeave        GroupEventType = "leave"
	GroupEventTypeRemoveMember GroupEventType = "remove_member"
	GroupEventTypeBlockMember  GroupEventType = "block_member"

	GroupEventTypeUpdateSetting GroupEventType = "update_setting"
	GroupEventTypeUpdate        GroupEventType = "update"
	GroupEventTypeNewLink       GroupEventType = "new_link"

	GroupEventTypeAddAdmin    GroupEventType = "add_admin"
	GroupEventTypeRemoveAdmin GroupEventType = "remove_admin"

	GroupEventTypeNewPinTopic     GroupEventType = "new_pin_topic"
	GroupEventTypeUpdatePinTopic  GroupEventType = "update_pin_topic"
	GroupEventTypeReorderPinTopic GroupEventType = "reorder_pin_topic"

	GroupEventTypeUpdateBoard GroupEventType = "update_board"
	GroupEventTypeRemoveBoard GroupEventType = "remove_board"

	GroupEventTypeUpdateTopic GroupEventType = "update_topic"
	GroupEventTypeUnpinTopic  GroupEventType = "unpin_topic"
	GroupEventTypeRemoveTopic GroupEventType = "remove_topic"

	GroupEventTypeAcceptRemind GroupEventType = "accept_remind"
	GroupEventTypeRejectRemind GroupEventType = "reject_remind"
	GroupEventTypeRemindTopic  GroupEventType = "remind_topic"

	GroupEventTypeUpdateAvatar GroupEventType = "update_avatar"

	GroupEventTypeUnknown GroupEventType = "unknown"
)

func ParseGroupEventType(s string) GroupEventType {
	switch s {
	case "join_request":
		return GroupEventTypeJoinRequest
	case "join":
		return GroupEventTypeJoin
	case "leave":
		return GroupEventTypeLeave
	case "remove_member":
		return GroupEventTypeRemoveMember
	case "block_member":
		return GroupEventTypeBlockMember
	case "update_setting":
		return GroupEventTypeUpdateSetting
	case "update":
		return GroupEventTypeUpdate
	case "new_link":
		return GroupEventTypeNewLink
	case "add_admin":
		return GroupEventTypeAddAdmin
	case "remove_admin":
		return GroupEventTypeRemoveAdmin
	case "new_pin_topic":
		return GroupEventTypeNewPinTopic
	case "update_pin_topic":
		return GroupEventTypeUpdatePinTopic
	case "reorder_pin_topic":
		return GroupEventTypeReorderPinTopic
	case "update_board":
		return GroupEventTypeUpdateBoard
	case "remove_board":
		return GroupEventTypeRemoveBoard
	case "update_topic":
		return GroupEventTypeUpdateTopic
	case "unpin_topic":
		return GroupEventTypeUnpinTopic
	case "remove_topic":
		return GroupEventTypeRemoveTopic
	case "accept_remind":
		return GroupEventTypeAcceptRemind
	case "reject_remind":
		return GroupEventTypeRejectRemind
	case "remind_topic":
		return GroupEventTypeRemindTopic
	case "update_avatar":
		return GroupEventTypeUpdateAvatar
	default:
		return GroupEventTypeUnknown
	}
}

type GroupEvent interface {
	IsSelf() bool
	Data() TGroupEvent
	Action() string
	Type() GroupEventType
	ThreadID() string
}

func NewGroupEvent(uid, action string, data TGroupEvent) GroupEvent {
	grEvent := ParseGroupEventType(action)
	isSelf := false

	switch grEvent {
	case GroupEventTypeJoin:
		isSelf = false
	case GroupEventTypeNewPinTopic, GroupEventTypeUnpinTopic, GroupEventTypeUpdatePinTopic:
		isSelf = data.(TGroupEventPinTopic).ActorID == uid
	case GroupEventTypeReorderPinTopic:
		isSelf = data.(TGroupEventReorderPinTopic).ActorID == uid
	case GroupEventTypeUpdateBoard, GroupEventTypeRemoveBoard:
		isSelf = data.(TGroupEventBoard).SourceID == uid
	case GroupEventTypeAcceptRemind, GroupEventTypeRejectRemind:
		isSelf = slices.Contains(data.(TGroupEventRemindRespond).UpdateMembers, uid)
	case GroupEventTypeRemindTopic:
		isSelf = data.(TGroupEventRemindTopic).CreatorID == uid
	default:
		baseData := data.(TGroupEventBase)
		if len(baseData.UpdateMembers) == 0 {
			isSelf = baseData.SourceID == uid
		} else {
			isSelf = slices.ContainsFunc(baseData.UpdateMembers,
				func(m GroupEventUpdateMember) bool {
					return m.ID == uid
				},
			)
		}
	}

	return groupEvent{
		typ:      grEvent,
		data:     data,
		act:      action,
		threadID: data.GroupID(),
		isSelf:   isSelf,
	}
}

type groupEvent struct {
	typ      GroupEventType
	data     TGroupEvent
	act      string
	threadID string
	isSelf   bool
}

func (e groupEvent) Type() GroupEventType { return e.typ }
func (e groupEvent) Data() TGroupEvent    { return e.data }
func (e groupEvent) Action() string       { return e.act }
func (e groupEvent) ThreadID() string     { return e.threadID }
func (e groupEvent) IsSelf() bool         { return e.isSelf }

type TGroupEvent interface {
	EventType() EventType
	GroupID() string
}

type TGroupEventBase struct {
	GID string `json:"groupId"`

	SubType       int                      `json:"subType"`
	CreatorID     string                   `json:"creatorId"`
	GroupName     string                   `json:"groupName"`
	SourceID      string                   `json:"sourceId"`
	UpdateMembers []GroupEventUpdateMember `json:"updateMembers"`
	GroupSetting  *GroupSetting            `json:"groupSetting"`
	GroupTopic    *GroupTopic              `json:"groupTopic"`
	Info          GroupEventGroupInfo      `json:"info"`
	ExtraData     GroupEventExtraData      `json:"extraData"`
	Time          string                   `json:"time"`
	Avt           *string                  `json:"avt"`
	FullAvt       *string                  `json:"fullAvt"`
	IsAdd         int                      `json:"isAdd"`
	HideGroupInfo int                      `json:"hideGroupInfo"`
	Version       string                   `json:"version"`
	GroupType     int                      `json:"groupType"`
	ClientID      *int                     `json:"clientId,omitempty"`
	ErrorMap      map[string]any           `json:"errorMap,omitempty"`
	E2EE          *int                     `json:"e2ee,omitempty"`
}

func (e TGroupEventBase) EventType() EventType { return EventTypeGroup }
func (e TGroupEventBase) GroupID() string      { return e.GID }

type TGroupEventJoinRequest struct {
	GID          string   `json:"groupId"`
	UIDs         []string `json:"uids"`
	TotalPending int      `json:"totalPending"`
	Time         string   `json:"time"`
}

func (e TGroupEventJoinRequest) EventType() EventType { return EventTypeGroup }
func (e TGroupEventJoinRequest) GroupID() string      { return e.GID }

type TGroupEventPinTopic struct {
	GID             string     `json:"groupId"`
	OldBoardVersion int        `json:"oldBoardVersion"`
	BoardVersion    int        `json:"boardVersion"`
	Topic           GroupTopic `json:"topic"`
	ActorID         string     `json:"actorId"`
}

func (e TGroupEventPinTopic) EventType() EventType { return EventTypeGroup }
func (e TGroupEventPinTopic) GroupID() string      { return e.GID }

type TGroupEventReorderPinTopic struct {
	GID             string `json:"groupId"`
	OldBoardVersion int    `json:"oldBoardVersion"`
	ActorID         string `json:"actorId"`
	Topics          []struct {
		TopicID   string `json:"topicId"`
		TopicType int    `json:"topicType"`
	} `json:"topics"`
	BoardVersion int `json:"boardVersion"`
	Topic        any `json:"topic"`
}

func (e TGroupEventReorderPinTopic) EventType() EventType { return EventTypeGroup }
func (e TGroupEventReorderPinTopic) GroupID() string      { return e.GID }

type TGroupEventBoard struct {
	GID        string `json:"groupId"`
	SourceID   string `json:"sourceId"`
	GroupName  string `json:"groupName"`
	GroupTopic any    `json:"groupTopic"` // (model.GroupTopic | model.ReminderGroup)
	CreatorID  string `json:"creatorId"`

	SubType       *int                     `json:"subType,omitempty"`
	UpdateMembers []GroupEventUpdateMember `json:"updateMembers,omitempty"`
	GroupSetting  *GroupSetting            `json:"groupSetting,omitempty"`
	Info          GroupEventGroupInfo      `json:"info,omitempty"`
	ExtraData     GroupEventExtraData      `json:"extraData,omitempty"`
	Time          *string                  `json:"time,omitempty"`
	Avt           *string                  `json:"avt,omitempty"`
	FullAvt       *string                  `json:"fullAvt,omitempty"`
	IsAdd         *int                     `json:"isAdd,omitempty"`
	HideGroupInfo *int                     `json:"hideGroupInfo,omitempty"`
	Version       *string                  `json:"version,omitempty"`
	GroupType     *int                     `json:"groupType,omitempty"`
}

func (e TGroupEventBoard) EventType() EventType { return EventTypeGroup }
func (e TGroupEventBoard) GroupID() string      { return e.GID }

type TGroupEventRemindRespond struct {
	GID           string   `json:"groupId"`
	TopicID       string   `json:"topicId"`
	UpdateMembers []string `json:"updateMembers"`
	Time          string   `json:"time"`
}

func (e TGroupEventRemindRespond) EventType() EventType { return EventTypeGroup }
func (e TGroupEventRemindRespond) GroupID() string      { return e.GID }

type TGroupEventRemindTopic struct {
	GID        string `json:"group_id"`
	Msg        string `json:"msg"`
	EditorID   string `json:"editorId"`
	Color      string `json:"color"`
	Emoji      string `json:"emoji"`
	CreatorID  string `json:"creatorId"`
	EditTime   int64  `json:"editTime"`
	Type       int    `json:"type"`
	Duration   int64  `json:"duration"`
	CreateTime int64  `json:"createTime"`
	Repeat     int    `json:"repeat"`
	StartTime  int64  `json:"startTime"`
	Time       int64  `json:"time"`
	RemindType int    `json:"remindType"`
}

func (e TGroupEventRemindTopic) EventType() EventType { return EventTypeGroup }
func (e TGroupEventRemindTopic) GroupID() string      { return e.GID }

type GroupEventUpdateMember struct {
	ID       string `json:"id"`
	DName    string `json:"dName"`
	Avatar   string `json:"avatar"`
	Type     int    `json:"type"`
	Avatar25 string `json:"avatar_25"`
}

type GroupEventGroupInfo struct {
	GroupLink       *string `json:"group_link,omitempty"`
	LinkExpiredTime *int64  `json:"link_expired_time,omitempty"`

	// ...
}

type GroupEventExtraData struct {
	FeatureID *int    `json:"featureId,omitempty"`
	Field     *string `json:"field,omitempty"`

	// ...
}

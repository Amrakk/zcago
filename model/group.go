package model

import "encoding/json"

type GroupTopicType int

const (
	GroupTopicNote    GroupTopicType = 0
	GroupTopicMessage GroupTopicType = 2
	GroupTopicPoll    GroupTopicType = 3
)

type GroupType int

const (
	GroupTypeGroup GroupType = iota + 1
	GroupTypeCommunity
)

type GroupSetting struct {
	BlockName        int `json:"blockName"`
	SignAdminMsg     int `json:"signAdminMsg"`
	AddMemberOnly    int `json:"addMemberOnly"`
	SetTopicOnly     int `json:"setTopicOnly"`
	EnableMsgHistory int `json:"enableMsgHistory"`
	JoinAppr         int `json:"joinAppr"`
	LockCreatePost   int `json:"lockCreatePost"`
	LockCreatePoll   int `json:"lockCreatePoll"`
	LockSendMsg      int `json:"lockSendMsg"`
	LockViewMember   int `json:"lockViewMember"`
	BannFeature      int `json:"bannFeature"`
	DirtyMedia       int `json:"dirtyMedia"`
	BanDuration      int `json:"banDuration"`
}

type GroupTopic struct {
	ID        string         `json:"id"`
	EditorID  string         `json:"editorId"`
	CreatorID string         `json:"creatorId"`
	Type      GroupTopicType `json:"type"`
	Color     int            `json:"color"`
	Emoji     string         `json:"emoji"`
	// JSON string, Unmarshal dynamically into specific structs
	//
	// Possible types:
	//   - GroupTopicNoteParams
	//   - GroupTopicTextMessageParams
	//   - GroupTopicFileMessageParams
	//   - GroupTopicPollParams
	//   - GroupTopicOtherParams
	Params     string `json:"params"`
	Repeat     int    `json:"repeat"`
	Action     int    `json:"action"`
	Duration   int64  `json:"duration"`
	CreateTime int64  `json:"createTime"`
	StartTime  int64  `json:"startTime"`
	EditTime   int64  `json:"editTime"`
}

type GroupInfo struct {
	GroupID        string         `json:"groupId"`
	Name           string         `json:"name"`
	Description    string         `json:"desc"`
	Type           GroupType      `json:"type"`
	CreatorID      string         `json:"creatorId"`
	Version        string         `json:"version"`
	Avatar         string         `json:"avt"`
	FullAvatar     string         `json:"fullAvt"`
	MemberIDs      []string       `json:"memberIds"`
	AdminIDs       []string       `json:"adminIds"`
	CurrentMembers []UserSummary  `json:"currentMems"`
	UpdateMembers  []any          `json:"updateMems"`
	Admins         []any          `json:"admins"`
	HasMoreMember  int            `json:"hasMoreMember"`
	SubType        int            `json:"subType"`
	TotalMember    int            `json:"totalMember"`
	MaxMember      int            `json:"maxMember"`
	Setting        GroupSetting   `json:"setting"`
	CreatedTime    int64          `json:"createdTime"`
	Visibility     int            `json:"visibility"`
	GlobalID       string         `json:"globalId"`
	E2EE           int            `json:"e2ee"` // 1: True, 0: False
	ExtraInfo      ExtraGroupInfo `json:"extraInfo"`
}

type ExtraGroupInfo struct {
	EnableMediaStore int `json:"enable_media_store"`
}

type GroupTopicNoteParams struct {
	ClientMsgID string `json:"client_msg_id"`
	GlobalMsgID string `json:"global_msg_id"`
	Title       string `json:"title"`
}

type GroupTopicTextMessageParams struct {
	SenderUID   string  `json:"senderUid"`
	SenderName  string  `json:"senderName"`
	ClientMsgID string  `json:"client_msg_id"`
	GlobalMsgID string  `json:"global_msg_id"`
	MsgType     int     `json:"msg_type"`
	Title       string  `json:"title"`
	Thumb       *string `json:"thumb,omitempty"`
}

type GroupTopicFileMessageParams struct {
	GroupTopicTextMessageParams
	Extra GroupTopicFileMessageExtra `json:"extra"`
}

func (p *GroupTopicFileMessageParams) UnmarshalJSON(data []byte) error {
	type alias GroupTopicFileMessageParams
	aux := &struct {
		Extra string `json:"extra"`
		*alias
	}{
		alias: (*alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Extra == "" {
		p.Extra = GroupTopicFileMessageExtra{}
		return nil
	}

	var extra GroupTopicFileMessageExtra
	if err := json.Unmarshal([]byte(aux.Extra), &extra); err != nil {
		return err
	}
	p.Extra = extra

	return nil
}

type GroupTopicFileMessageExtra struct {
	FileSize    string `json:"fileSize"`
	Checksum    string `json:"checksum"`
	ChecksumSha any    `json:"checksumSha"`
	FileExt     string `json:"fileExt"`
	FData       string `json:"fdata"`
	FType       int    `json:"fType"`
}

type GroupTopicPollParams struct {
	PollID int    `json:"pollId"`
	Title  string `json:"title"`
}

type GroupTopicOtherParams map[string]any

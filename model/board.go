package model

import (
	"encoding/json"
)

type BoardType int

const (
	BoardTypeNote BoardType = iota + 1
	BoardTypePinnedMessage
	BoardTypePoll
)

type BoardData interface {
	GetBoardType() BoardType
}

type PollDetail struct {
	PollID            int          `json:"poll_id"`
	Creator           string       `json:"creator"`
	Question          string       `json:"question"`
	PollType          int          `json:"poll_type"`
	Options           []PollOption `json:"options"`
	Joined            bool         `json:"joined"`
	Closed            bool         `json:"closed"`
	AllowMultiChoices bool         `json:"allow_multi_choices"`
	AllowAddNewOption bool         `json:"allow_add_new_option"`
	IsAnonymous       bool         `json:"is_anonymous"`
	IsHideVotePreview bool         `json:"is_hide_vote_preview"`
	NumVote           int          `json:"num_vote"`
	CreatedTime       int64        `json:"created_time"`
	UpdatedTime       int64        `json:"updated_time"`
	ExpiredTime       int64        `json:"expired_time"`
}

func (p *PollDetail) GetBoardType() BoardType { return BoardTypePoll }

type PollOption struct {
	OptionId int      `json:"option_id"`
	Content  string   `json:"content"`
	Votes    int      `json:"votes"`
	Voted    bool     `json:"voted"`
	Voters   []string `json:"voters"`
}

type NoteParams struct {
	Title string  `json:"title"`
	Extra *string `json:"extra,omitempty"`
}

type NoteDetail struct {
	ID         string     `json:"id"`
	CreatorID  string     `json:"creatorId"`
	EditorID   string     `json:"editorId"`
	Type       int        `json:"type"`
	Color      int        `json:"color"`
	Emoji      string     `json:"emoji"`
	Params     NoteParams `json:"params"`
	Repeat     int        `json:"repeat"`
	Duration   int        `json:"duration"`
	StartTime  int64      `json:"startTime"`
	CreateTime int64      `json:"createTime"`
	EditTime   int64      `json:"editTime"`
}

func (n *NoteDetail) GetBoardType() BoardType { return BoardTypeNote }

func (n *NoteDetail) UnmarshalJSON(data []byte) error {
	type alias NoteDetail
	aux := &struct {
		Params json.RawMessage `json:"params"`
		*alias
	}{
		alias: (*alias)(n),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Params) > 0 && string(aux.Params) != "null" {
		var params NoteParams
		if err := json.Unmarshal(aux.Params, &params); err == nil {
			n.Params = params
		}
	}

	return nil
}

type PinnedMessageDetail struct {
	ID         string                    `json:"id"`
	CreatorID  string                    `json:"creatorId"`
	EditorID   string                    `json:"editorId"`
	Type       int                       `json:"type"`
	Color      int                       `json:"color"`
	Emoji      string                    `json:"emoji"`
	Params     PinnedMessageDetailParams `json:"params"`
	Repeat     int                       `json:"repeat"`
	Duration   int                       `json:"duration"`
	StartTime  int64                     `json:"startTime"`
	CreateTime int64                     `json:"createTime"`
	EditTime   int64                     `json:"editTime"`
}

func (pm *PinnedMessageDetail) GetBoardType() BoardType { return BoardTypePinnedMessage }

type PinnedMessageDetailParams struct {
	SenderUID   string `json:"senderUid"`
	SenderName  string `json:"senderName"`
	ClientMsgID string `json:"client_msg_id"`
	Thumb       string `json:"thumb"`
	GlobalMsgID string `json:"global_msg_id"`
	MsgType     int    `json:"msg_type"`
	Title       string `json:"title"`
}

func (p *PinnedMessageDetailParams) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	type alias PinnedMessageDetailParams
	var a alias

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		data = []byte(s)
	}

	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*p = PinnedMessageDetailParams(a)

	return nil
}

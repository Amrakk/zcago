package model

type TUndoContent struct {
	GlobalMsgID int64 `json:"globalMsgId"`
	CliMsgID    int64 `json:"cliMsgId"`
	DeleteMsg   int   `json:"deleteMsg"`
	SrcID       int64 `json:"srcId"`
	DestID      int64 `json:"destId"`
}

type TUndo struct {
	ActionID  string       `json:"actionId"`
	MsgID     string       `json:"msgId"`
	CliMsgID  string       `json:"cliMsgId"`
	MsgType   string       `json:"msgType"`
	UIDFrom   string       `json:"uidFrom"`
	IDTo      string       `json:"idTo"`
	DName     string       `json:"dName"`
	TS        string       `json:"ts"`
	Status    int          `json:"status"`
	Content   TUndoContent `json:"content"`
	Notify    string       `json:"notify"`
	TTL       int          `json:"ttl"`
	UserID    string       `json:"userId"`
	UIN       string       `json:"uin"`
	Cmd       int          `json:"cmd"`
	ST        int          `json:"st"`
	AT        int          `json:"at"`
	RealMsgID string       `json:"realMsgId"`
}

type Undo struct {
	Data     TUndo
	ThreadID string
	IsSelf   bool
	IsGroup  bool
}

func NewUndo(uid string, data TUndo, isGroup bool) Undo {
	u := Undo{
		Data:    data,
		IsSelf:  data.UIDFrom == "0",
		IsGroup: isGroup,
	}

	if isGroup || data.UIDFrom == "0" {
		u.ThreadID = data.IDTo
	} else {
		u.ThreadID = data.UIDFrom
	}

	if u.Data.IDTo == "0" {
		u.Data.IDTo = uid
	}
	if u.Data.UIDFrom == "0" {
		u.Data.UIDFrom = uid
	}

	return u
}

package events

import (
	"encoding/json"

	"github.com/Amrakk/zcago/model"
)

type GroupMessageEventData struct {
	GroupMsgs []groupMessageOrUndo `json:"groupMsgs"`
}

type groupMessageOrUndo struct {
	Message *model.TGroupMessage
	Undo    *model.TUndo
}

func (m *groupMessageOrUndo) UnmarshalJSON(data []byte) error {
	var tu model.TUndo
	if err := json.Unmarshal(data, &tu); err == nil && tu.MsgID != "" && tu.MsgType == "chat.undo" {
		m.Undo = &tu
		return nil
	}
	var tm model.TGroupMessage
	if err := json.Unmarshal(data, &tm); err == nil && tm.MsgID != "" {
		m.Message = &tm
		return nil
	}

	return nil
}

package events

import (
	"encoding/json"

	"github.com/Amrakk/zcago/model"
)

type MessageEventData struct {
	Msgs []messageOrUndo `json:"msgs"`
}

type messageOrUndo struct {
	Message *model.TMessage
	Undo    *model.TUndo
}

func (m *messageOrUndo) UnmarshalJSON(data []byte) error {
	var tu model.TUndo
	if err := json.Unmarshal(data, &tu); err == nil && tu.MsgID != "" && tu.MsgType == "chat.undo" {
		m.Undo = &tu
		return nil
	}
	var tm model.TMessage
	if err := json.Unmarshal(data, &tm); err == nil && tm.MsgID != "" {
		m.Message = &tm
		return nil
	}

	return nil
}

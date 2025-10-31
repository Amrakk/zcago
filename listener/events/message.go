package events

import (
	"encoding/json"

	"github.com/Amrakk/zcago/errs"
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

func (m messageOrUndo) MarshalJSON() ([]byte, error) {
	if m.Message != nil {
		return json.Marshal(m.Message)
	}
	if m.Undo != nil {
		return json.Marshal(m.Undo)
	}
	return nil, errs.NewZCA("both Message and Undo are nil", "messageOrUndo.MarshalJSON")
}

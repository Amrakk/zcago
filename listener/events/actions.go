package events

import (
	"encoding/json"
	"strings"

	"github.com/Amrakk/zcago/model"
)

type ActionEventData struct {
	Actions []actionItem `json:"actions"`
}

type actionItem struct {
	ActionType string     `json:"act_type"`
	Action     string     `json:"act"`
	Data       actionData `json:"data"`
}

type actionData struct {
	Typing      model.TTyping
	GroupTyping model.TGroupTyping
}

func (m *actionData) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil || s == "" {
			return err
		}

		if !strings.HasPrefix(s, "{") {
			s = "{" + s
		}
		if !strings.HasSuffix(s, "}") {
			s += "}"
		}

		data = []byte(s)
	}

	var gt model.TGroupTyping
	if err := json.Unmarshal(data, &gt); err == nil && gt.GID != "" {
		m.GroupTyping = gt
		return nil
	}
	var t model.TTyping
	if err := json.Unmarshal(data, &t); err == nil {
		m.Typing = t
		return nil
	}

	return nil
}

package events

import (
	"encoding/json"
)

type ControlEventData struct {
	Controls []controlItem `json:"controls"`
}

type controlItem struct {
	Content controlContent `json:"content"`
}

type controlContent struct {
	ActionType string         `json:"act_type"`
	Action     string         `json:"act"`
	Data       controlPayload `json:"data"`

	FileID *int64 `json:"fileId,omitempty"`
}

type controlPayload struct {
	UploadAttachment *uploadFileInfo `json:"uploadAttachment,omitempty"`
}

type uploadFileInfo struct {
	URL string `json:"url"`
}

func (d *ControlEventData) UnmarshalJSON(data []byte) error {
	type alias ControlEventData
	var tmp alias

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*d = ControlEventData(tmp)
	return nil
}

func (m *controlPayload) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		data = []byte(s)
	}

	var ul uploadFileInfo
	if err := json.Unmarshal(data, &ul); err == nil {
		m.UploadAttachment = &ul
		return nil
	}

	return nil
}

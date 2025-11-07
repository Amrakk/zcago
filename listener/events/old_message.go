package events

import (
	"github.com/Amrakk/zcago/model"
)

type OldMessagesEventData struct {
	Msgs      []model.TMessage `json:"msgs"`
	GroupMsgs []model.TMessage `json:"groupMsgs"`
}

package events

import (
	"github.com/amrakk/zcago/model"
)

type OldMessagesEventData struct {
	Msgs      []model.TMessage      `json:"msgs"`
	GroupMsgs []model.TGroupMessage `json:"groupMsgs"`
}

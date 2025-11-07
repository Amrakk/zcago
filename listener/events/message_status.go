package events

import (
	"github.com/Amrakk/zcago/model"
)

type MessageStatusEventData struct {
	DeliveredMessages []model.TDeliveredMessage `json:"delivereds"`
	SeenMessages      []model.TUserSeenMessage  `json:"seens"`
}

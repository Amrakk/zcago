package events

import "github.com/Amrakk/zcago/model"

type GroupMessageStatusEventData struct {
	DeliveredMessages []model.TGroupDeliveredMessage `json:"delivereds"`
	SeenMessages      []model.TGroupSeenMessage      `json:"groupSeens"`
}

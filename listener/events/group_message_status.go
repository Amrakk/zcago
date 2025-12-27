package events

import "github.com/amrakk/zcago/model"

type GroupMessageStatusEventData struct {
	DeliveredMessages []model.TGroupDeliveredMessage `json:"delivereds"`
	SeenMessages      []model.TGroupSeenMessage      `json:"groupSeens"`
}

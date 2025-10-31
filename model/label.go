package model

type LabelData struct {
	ID            int      `json:"id"`
	Text          string   `json:"text"`
	TextKey       string   `json:"textKey"`
	Conversations []string `json:"conversations"`
	Color         string   `json:"color"`
	Offset        int      `json:"offset"`
	Emoji         string   `json:"emoji"`
	CreateTime    int64    `json:"createTime"`
}

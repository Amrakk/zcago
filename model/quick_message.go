package model

type QuickMessageType int

const (
	QuickMessageTypeText QuickMessageType = iota
	QuickMessageTypeMedia
)

type QuickMessage struct {
	ID           int                 `json:"id"`
	Keyword      string              `json:"keyword"`
	Type         QuickMessageType    `json:"type"`
	CreatedTime  int64               `json:"createdTime"`
	LastModified int64               `json:"lastModified"`
	Message      QuickMessageContent `json:"message"`
	Media        *QuickMessageMedia  `json:"media"`
}

type QuickMessageContent struct {
	Title  string  `json:"title"`
	Params *string `json:"params"`
}

type QuickMessageMedia struct {
	Items []QuickMessageMediaItem `json:"items"`
}

type QuickMessageMediaItem struct {
	Type         int    `json:"type"`
	PhotoID      int    `json:"photoId"`
	Title        string `json:"title"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	PreviewThumb string `json:"previewThumb"`
	RawURL       string `json:"rawUrl"`
	ThumbURL     string `json:"thumbUrl"`
	NormalURL    string `json:"normalUrl"`
	HDURL        string `json:"hdUrl"`
}

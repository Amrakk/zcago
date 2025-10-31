package model

type StickerBasic struct {
	Type      int `json:"type"`
	CateID    int `json:"cate_id"`
	StickerID int `json:"sticker_id"`
}
type StickerSuggestions struct {
	SuggSticker []StickerBasic `json:"sugg_sticker"`
	SuggGuggy   []StickerBasic `json:"sugg_guggy"`
	SuggGif     []StickerBasic `json:"sugg_gif"`
}

type StickerDetail struct {
	ID               int    `json:"id"`
	CateID           int    `json:"cateId"`
	Type             int    `json:"type"`
	Text             string `json:"text"`
	URI              string `json:"uri"`
	FKey             int    `json:"fkey"`
	Status           int    `json:"status"`
	StickerURL       string `json:"stickerUrl"`
	StickerSpriteURL string `json:"stickerSpriteUrl"`
	StickerWebpURL   any    `json:"stickerWebpUrl"`
	TotalFrames      int    `json:"totalFrames"`
	Duration         int    `json:"duration"`
	EffectID         int    `json:"effectId"`
	Checksum         string `json:"checksum"`
	Ext              int    `json:"ext"`
	Source           int    `json:"source"`
	Fss              any    `json:"fss"`
	FssInfo          any    `json:"fssInfo"`
	Version          int    `json:"version"`
	ExtInfo          any    `json:"extInfo"`
}

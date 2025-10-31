package model

type ErrMap map[string]struct {
	Err int `json:"err"`
}

type OffsetPaginationOptions struct {
	Count int
	Page  int
}

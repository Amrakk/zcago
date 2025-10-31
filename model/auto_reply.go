package model

type AutoReplyScope int

const (
	AutoReplyScopeEveryone AutoReplyScope = iota
	AutoReplyScopeStranger
	AutoReplyScopeSpecificFriends
	AutoReplyScopeFriendsExcept
)

type AutoReplyItem struct {
	ID           int            `json:"id"`
	UIDs         []string       `json:"uids"`
	OwnerId      int            `json:"ownerId"`
	Content      string         `json:"content"`
	Weight       int            `json:"weight"`
	Enable       bool           `json:"enable"`
	Scope        AutoReplyScope `json:"scope"`
	StartTime    int64          `json:"startTime"`
	EndTime      int64          `json:"endTime"`
	Recurrence   []string       `json:"recurrence"`
	ModifiedTime int64          `json:"modifiedTime"`
	CreatedTime  int64          `json:"createdTime"`
}

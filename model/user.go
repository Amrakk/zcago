package model

type Gender int

const (
	Male   Gender = 0
	Female Gender = 1
)

type User struct {
	UserId         string           `json:"userId"`
	Username       string           `json:"username"`
	DisplayName    string           `json:"displayName"`
	ZaloName       string           `json:"zaloName"`
	Avatar         string           `json:"avatar"`
	Bgavatar       string           `json:"bgavatar"`
	Cover          string           `json:"cover"`
	Gender         Gender           `json:"gender"`
	Dob            int64            `json:"dob"`
	Sdob           string           `json:"sdob"`
	Status         string           `json:"status"`
	PhoneNumber    string           `json:"phoneNumber"`
	IsFr           int              `json:"isFr"`
	IsBlocked      int              `json:"isBlocked"`
	LastActionTime int64            `json:"lastActionTime"`
	LastUpdateTime int64            `json:"lastUpdateTime"`
	IsActive       int              `json:"isActive"`
	Key            int              `json:"key"`
	Type           int              `json:"type"`
	IsActivePC     int              `json:"isActivePC"`
	IsActiveWeb    int              `json:"isActiveWeb"`
	IsValid        int              `json:"isValid"`
	UserKey        string           `json:"userKey"`
	AccountStatus  int              `json:"accountStatus"`
	OAInfo         interface{}      `json:"oaInfo"`
	UserMode       int              `json:"user_mode"`
	GlobalId       string           `json:"globalId"`
	BizPkg         ZBusinessPackage `json:"bizPkg"`
	CreatedTs      int64            `json:"createdTs"`
	OAStatus       interface{}      `json:"oa_status"`
}

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
	OAInfo         any              `json:"oaInfo"`
	UserMode       int              `json:"user_mode"`
	GlobalId       string           `json:"globalId"`
	BizPkg         ZBusinessPackage `json:"bizPkg"`
	CreatedTs      int64            `json:"createdTs"`
	OAStatus       any              `json:"oa_status"`
}
type UserSetting struct {
	AddFriendViaContact      int  `json:"add_friend_via_contact"`
	DisplayOnRecommendFriend int  `json:"display_on_recommend_friend"`
	AddFriendViaGroup        int  `json:"add_friend_via_group"`
	AddFriendViaQR           int  `json:"add_friend_via_qr"`
	QuickMessageStatus       int  `json:"quick_message_status"`
	ShowOnlineStatus         bool `json:"show_online_status"`
	AcceptStrangerCall       int  `json:"accept_stranger_call"`
	ArchivedChatStatus       int  `json:"archived_chat_status"`
	ReceiveMessage           int  `json:"receive_message"`
	AddFriendViaPhone        int  `json:"add_friend_via_phone"`
	DisplaySeenStatus        int  `json:"display_seen_status"`
	ViewBirthday             int  `json:"view_birthday"`
	Setting2FAStatus         int  `json:"setting_2FA_status"`
}

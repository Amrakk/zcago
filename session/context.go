package session

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/Amrakk/zcago/internal/jsonx"
)

type BaseContext struct {
	UID       string
	IMEI      string
	Cookie    http.CookieJar
	UserAgent string

	Options    Options
	APIType    uint
	APIVersion uint
}

type Context struct {
	UID       string
	IMEI      string
	Cookie    http.CookieJar
	UserAgent string
	SecretKey string

	Language  string
	Settings  Settings
	LoginInfo *LoginInfo
	ExtraVer  ExtraVer

	WebsocketURL []string

	UploadCallbacks *CallbacksMap
	Options         Options
	APIType         uint
	APIVersion      uint
}

func NewContext(apiType, apiVersion uint) *Context {
	jar, _ := cookiejar.New(nil)
	return &Context{
		APIType:         apiType,
		APIVersion:      apiVersion,
		Cookie:          jar,
		UploadCallbacks: NewCallbacksMap(),
	}
}

func (c *Context) IsLogging() bool {
	return c.Options.Logging
}

func (c *Context) ZpwServiceMap() *ZpwServiceMap {
	if c == nil || c.LoginInfo == nil {
		return nil
	}
	return &c.LoginInfo.ZpwServiceMapV3
}

func (c *Context) ZpwWebsocket() []string {
	return c.LoginInfo.ZpwWebsocket
}

type Settings struct {
	Features  Features          `json:"features"`
	Keepalive KeepaliveSettings `json:"keepalive"`
}

type ShareFileSettings struct {
	BigFileDomainList     []string `json:"big_file_domain_list"`
	MaxSizeShareFileV2    uint     `json:"max_size_share_file_v2"`
	MaxSizeShareFileV3    uint     `json:"max_size_share_file_v3"`
	FileUploadShowIcon1GB bool     `json:"file_upload_show_icon_1GB"`
	RestrictedExt         string   `json:"restricted_ext"`
	NextFileTime          uint     `json:"next_file_time"`
	MaxFile               uint     `json:"max_file"`
	MaxSizePhoto          uint     `json:"max_size_photo"`
	MaxSizeShareFile      uint     `json:"max_size_share_file"`
	MaxSizeResizePhoto    uint     `json:"max_size_resize_photo"`
	MaxSizeGif            uint     `json:"max_size_gif"`
	MaxSizeOriginalPhoto  uint     `json:"max_size_original_photo"`
	ChunkSizeFile         uint     `json:"chunk_size_file"`
	RestrictedExtFile     []string `json:"restricted_ext_file"`
}

type SocketSettings struct {
	RotateErrorCodes []int                        `json:"rotate_error_codes"`
	Retries          map[string]SocketRetryConfig `json:"retries"`
	Debug            SocketDebug                  `json:"debug"`
	PingInterval     uint                         `json:"ping_interval"`
	ResetEndpoint    uint                         `json:"reset_endpoint"`
	QueueCtrlAction  QueueCtrlActionIDMap         `json:"queue_ctrl_actionid_map"`
	CloseAndRetry    []int                        `json:"close_and_retry_codes"`
	MaxMsgSize       uint                         `json:"max_msg_size"`
	EnableCtrlSocket bool                         `json:"enable_ctrl_socket"`
	ReconnectAfterFB bool                         `json:"reconnect_after_fallback"`
	EnableChatSocket bool                         `json:"enable_chat_socket"`
	SubmitWssLog     bool                         `json:"submit_wss_log"`
	DisableLP        bool                         `json:"disable_lp"`
	OfflineMonitor   OfflineMonitor               `json:"offline_monitor"`
}

type SocketRetryConfig struct {
	Max   *uint                 `json:"max,omitempty"`
	Times jsonx.OneOrMany[uint] `json:"times"`
}

type SocketDebug struct {
	Enable bool `json:"enable"`
}

type QueueCtrlActionIDMap struct {
	CMD_611_0 string `json:"611_0"`
	CMD_610_1 string `json:"610_1"`
	CMD_610_0 string `json:"610_0"`
	CMD_603_0 string `json:"603_0"`
	CMD_611_1 string `json:"611_1"`
}

type OfflineMonitor struct {
	Enable bool `json:"enable"`
}

type Features struct {
	ShareFile ShareFileSettings `json:"sharefile"`
	Socket    SocketSettings    `json:"socket"`
}

type KeepaliveSettings struct {
	AlwaysKeepalive   bool `json:"alway_keepalive"`
	KeepaliveDuration uint `json:"keepalive_duration"`
	TimeDeactive      uint `json:"time_deactive"`
}

type ExtraVer struct {
	Phonebook              uint   `json:"phonebook"`
	ConvLabel              string `json:"conv_label"`
	Friend                 string `json:"friend"`
	VerStickerGiphySuggest uint   `json:"ver_sticker_giphy_suggest"`
	VerGiphyCate           uint   `json:"ver_giphy_cate"`
	Alias                  string `json:"alias"`
	VerStickerCateList     uint   `json:"ver_sticker_cate_list"`
	BlockFriend            string `json:"block_friend"`
}

type ZpwServiceMap = ZpwServiceMapV3

type LoginInfo struct {
	UID         string `json:"uid"`
	ZPWEnk      string `json:"zpw_enk"`
	HasPCClient uint   `json:"haspcclient"`
	PublicIP    string `json:"public_ip"`
	Language    string `json:"language"`
	Send2meID   string `json:"send2me_id"`

	ZpwWebsocket    []string        `json:"zpw_ws"`
	ZpwServiceMapV3 ZpwServiceMapV3 `json:"zpw_service_map_v3"`
}

type ZpwServiceMapV3 struct {
	OtherContact       []string `json:"other_contact"`
	ChatE2E            []string `json:"chat_e2e"`
	Workspace          []string `json:"workspace"`
	Catalog            []string `json:"catalog"`
	Boards             []string `json:"boards"`
	DownloadStickerUrl []string `json:"download_sticker_url"`
	SpContact          []string `json:"sp_contact"`
	ZcloudUpFile       []string `json:"zcloud_up_file"`
	MediaStoreSend2me  []string `json:"media_store_send2me"`
	PushAct            []string `json:"push_act"`
	Aext               []string `json:"aext"`
	Zfamily            []string `json:"zfamily"`
	GroupPoll          []string `json:"group_poll"`
	GroupCloudMessage  []string `json:"group_cloud_message"`
	MediaStore         []string `json:"media_store"`
	File               []string `json:"file"`
	AutoReply          []string `json:"auto_reply"`
	SyncAction         []string `json:"sync_action"`
	FriendLan          []string `json:"friend_lan"`
	Friend             []string `json:"friend"`
	Alias              []string `json:"alias"`
	Zimsg              []string `json:"zimsg"`
	GroupBoard         []string `json:"group_board"`
	Conversation       []string `json:"conversation"`
	Group              []string `json:"group"`
	FallbackLP         []string `json:"fallback_LP"`
	FriendBoard        []string `json:"friend_board"`
	UpFile             []string `json:"up_file"`
	Zavi               []string `json:"zavi"`
	Reaction           []string `json:"reaction"`
	VoiceCall          []string `json:"voice_call"`
	Profile            []string `json:"profile"`
	Sticker            []string `json:"sticker"`
	Label              []string `json:"label"`
	Consent            []string `json:"consent"`
	Zcloud             []string `json:"zcloud"`
	Chat               []string `json:"chat"`
	TodoUrl            []string `json:"todoUrl"`
	RecentSearch       []string `json:"recent_search"`
	GroupE2E           []string `json:"group_e2e"`
	QuickMessage       []string `json:"quick_message"`
}

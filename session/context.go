package session

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/Amrakk/zcago/internal/jsonx"
)

type Context interface {
	UID() string
	IMEI() string
	UserAgent() string
	Language() string

	APIType() uint
	APIVersion() uint

	Options() OptionsSnapshot
	IsLogging() bool
	LogLevel() uint8
	Client() *http.Client
	CheckUpdate() bool

	// Cookies returns a snapshot of cookies from the jar.
	//
	// Behavior:
	//   - If no domains are specified, it returns cookies for the default domains:
	//     "chat.zalo.me" and "zalo.me".
	//   - If one or more domains are provided, it returns cookies only for those
	//     domains.
	//
	// All returned cookies are copies; modifying them does not affect the jar.
	Cookies(domains ...string) []*http.Cookie
	SecretKey() string
	LoginInfo() *LoginInfo
	Settings() *Settings
	ExtraVer() *ExtraVer

	ZPWServiceMap() *ZpwServiceMap
	ZPWWebsocket() []string
}

type MutableContext interface {
	Context

	SealLogin(seal Seal) // one-shot finalization
	CookieJar() http.CookieJar
	SetCookieJar(j http.CookieJar)
	AddCookies(u *url.URL, cookies []*http.Cookie)

	SetIMEI(imei string)
	SetUserAgent(ua string)
	SetLanguage(lang string)
	AsReadOnly() Context
}

type OptionsSnapshot struct {
	SelfListen          bool
	CheckUpdate         bool
	Logging             bool
	LogLevel            uint8
	APIType             uint
	APIVersion          uint
	Client              *http.Client
	ImageMetadataGetter ImageMetadataGetter
}

type Seal struct {
	UID       string
	IMEI      string
	UserAgent string
	Language  string

	SecretKey string
	LoginInfo *LoginInfo
	Settings  *Settings
	ExtraVer  *ExtraVer
	Cookie    http.CookieJar
}

type contextImpl struct {
	uid       string
	imei      string
	userAgent string
	language  string

	apiType    uint
	apiVersion uint

	opts OptionsSnapshot

	secretKey string
	loginInfo *LoginInfo
	settings  *Settings
	extraVer  *ExtraVer
	cookie    http.CookieJar

	uploadCallbacks *CallbacksMap

	mu sync.RWMutex
}

func NewContext(optFns ...Option) *contextImpl {
	cfg := defaultOptions()
	for _, fn := range optFns {
		if fn != nil {
			fn(&cfg)
		}
	}

	jar, _ := cookiejar.New(nil)

	return &contextImpl{
		apiType:    cfg.apiType,
		apiVersion: cfg.apiVersion,
		opts: OptionsSnapshot{
			SelfListen:          cfg.selfListen,
			CheckUpdate:         cfg.checkUpdate,
			Logging:             cfg.logging,
			LogLevel:            cfg.logLevel,
			APIType:             cfg.apiType,
			APIVersion:          cfg.apiVersion,
			Client:              cfg.client,
			ImageMetadataGetter: cfg.imageMetadataGetter,
		},
		cookie:          jar,
		uploadCallbacks: NewCallbacksMap(),
		language:        "vi",
	}
}

var _ Context = (*contextImpl)(nil)
var _ MutableContext = (*contextImpl)(nil)

//
// Mutators
//

func (c *contextImpl) SealLogin(s Seal) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if s.Cookie != nil {
		c.cookie = s.Cookie
	}
	c.uid = s.UID
	c.imei = s.IMEI
	if s.UserAgent != "" {
		c.userAgent = s.UserAgent
	}
	if s.Language != "" {
		c.language = s.Language
	}

	c.secretKey = s.SecretKey
	c.loginInfo = s.LoginInfo
	c.settings = s.Settings
	c.extraVer = s.ExtraVer

}

func (c *contextImpl) CookieJar() http.CookieJar {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cookie
}

func (c *contextImpl) SetCookieJar(j http.CookieJar) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cookie = j
}

func (c *contextImpl) SetIMEI(imei string) {
	c.mu.Lock()
	c.imei = imei
	c.mu.Unlock()
}

func (c *contextImpl) SetUserAgent(ua string) {
	c.mu.Lock()
	c.userAgent = ua
	c.mu.Unlock()
}

func (c *contextImpl) SetLanguage(lang string) {
	if lang == "" {
		return
	}
	c.mu.Lock()
	c.language = lang
	c.mu.Unlock()
}

func (c *contextImpl) AddCookies(u *url.URL, cookies []*http.Cookie) {
	c.mu.RLock()
	jar := c.cookie
	c.mu.RUnlock()
	if jar != nil && u != nil && len(cookies) > 0 {
		jar.SetCookies(u, cookies)
	}
}
func (c *contextImpl) AsReadOnly() Context { return c }

//
// Getters
//

func (c *contextImpl) UID() string              { c.mu.RLock(); defer c.mu.RUnlock(); return c.uid }
func (c *contextImpl) IMEI() string             { c.mu.RLock(); defer c.mu.RUnlock(); return c.imei }
func (c *contextImpl) UserAgent() string        { c.mu.RLock(); defer c.mu.RUnlock(); return c.userAgent }
func (c *contextImpl) Language() string         { c.mu.RLock(); defer c.mu.RUnlock(); return c.language }
func (c *contextImpl) APIType() uint            { return c.apiType }
func (c *contextImpl) APIVersion() uint         { return c.apiVersion }
func (c *contextImpl) Options() OptionsSnapshot { return c.opts }
func (c *contextImpl) Client() *http.Client     { return c.opts.Client }
func (c *contextImpl) IsLogging() bool          { return c.opts.Logging }
func (c *contextImpl) LogLevel() uint8          { return c.opts.LogLevel }
func (c *contextImpl) CheckUpdate() bool        { return c.opts.CheckUpdate }

func (c *contextImpl) Cookies(domains ...string) []*http.Cookie {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cookie == nil {
		return nil
	}

	targets := domains
	if len(targets) == 0 {
		targets = []string{"zalo.me", "chat.zalo.me"}
	}

	type key struct {
		Name string
		Path string
	}

	trimDotLower := func(s string) string {
		s = strings.ToLower(s)
		if strings.HasPrefix(s, ".") {
			return s[1:]
		}
		return s
	}

	// Smaller rank => more parent-like
	domainRank := func(d string) int {
		d = trimDotLower(d)
		n := 0
		for _, p := range strings.Split(d, ".") {
			if p != "" {
				n++
			}
		}
		return n
	}

	cloneCookie := func(src *http.Cookie) *http.Cookie {
		if src == nil {
			return nil
		}
		cp := &http.Cookie{
			Name:       src.Name,
			Value:      src.Value,
			Path:       src.Path,
			Domain:     src.Domain,
			Expires:    src.Expires,
			RawExpires: src.RawExpires,
			MaxAge:     src.MaxAge,
			Secure:     src.Secure,
			HttpOnly:   src.HttpOnly,
			SameSite:   src.SameSite,
			Raw:        src.Raw,
			Unparsed:   src.Unparsed,
		}
		return cp
	}

	prefer := func(a, b *http.Cookie) *http.Cookie {
		if a == nil {
			return b
		}
		if b == nil {
			return a
		}
		ra := domainRank(a.Domain)
		rb := domainRank(b.Domain)
		if ra != rb {
			if ra < rb {
				return a
			}
			return b
		}
		if a.Expires.After(b.Expires) {
			return a
		}
		if b.Expires.After(a.Expires) {
			return b
		}
		return a
	}

	best := make(map[key]*http.Cookie)

	for _, host := range targets {
		u := &url.URL{Scheme: "https", Host: host, Path: "/"}
		for _, ck := range c.cookie.Cookies(u) {
			if ck == nil || ck.Name == "" {
				continue
			}
			k := key{Name: ck.Name, Path: ck.Path}
			cl := cloneCookie(ck)
			if cur, ok := best[k]; ok {
				best[k] = prefer(cur, cl)
			} else {
				best[k] = cl
			}
		}
	}

	out := make([]*http.Cookie, 0, len(best))
	for _, v := range best {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		di := trimDotLower(out[i].Domain)
		dj := trimDotLower(out[j].Domain)
		if di != dj {
			return di < dj
		}
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Name < out[j].Name
	})

	return out
}

func (c *contextImpl) SecretKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.secretKey
}
func (c *contextImpl) LoginInfo() *LoginInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loginInfo
}
func (c *contextImpl) Settings() *Settings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings
}
func (c *contextImpl) ExtraVer() *ExtraVer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.extraVer
}

func (c *contextImpl) ZPWServiceMap() *ZpwServiceMap {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.loginInfo == nil {
		return nil
	}
	return &c.loginInfo.ZpwServiceMapV3
}
func (c *contextImpl) ZPWWebsocket() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.loginInfo == nil {
		return nil
	}
	return c.loginInfo.ZpwWebsocket
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

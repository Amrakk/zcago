package session

import (
	"net/http"
	"net/url"
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

	Client() *http.Client
	SetIMEI(imei string)
	SetUserAgent(ua string)
	SetLanguage(lang string)

	CookieJar() http.CookieJar
	SetCookieJar(j http.CookieJar)
	AddCookies(u *url.URL, cookies []*http.Cookie)

	AsReadOnly() Context
}

func NewContext(optFns ...Option) MutableContext {
	impl := newContextImpl(optFns...)
	return impl
}

var _ Context = (*contextImpl)(nil)
var _ MutableContext = (*contextImpl)(nil)

package session

import (
	"net/http"
	"net/url"
	"time"

	"github.com/Amrakk/zcago/model"
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
	GetImageMetadata(path string) (model.AttachmentMetadata, string, error)

	CookieJar() http.CookieJar
	SecretKey() SecretKey
	LoginInfo() *LoginInfo
	Settings() *Settings
	ExtraVer() *ExtraVer
	UploadCallback() *CallbacksMap

	ZPWWebsocket() []string
	WSPingInterval() time.Duration

	ZPWServiceMap() *ZpwServiceMap
	GetZpwService(service string) []string
}

type MutableContext interface {
	Context

	SealLogin(seal Seal) // one-shot finalization

	Client() *http.Client
	Proxy() func(*http.Request) (*url.URL, error)

	SetIMEI(imei string)
	SetUserAgent(ua string)
	SetLanguage(lang string)

	SetCookieJar(j http.CookieJar)

	AsReadOnly() Context
}

func NewContext(optFns ...Option) MutableContext {
	return newContextImpl(optFns...)
}

var (
	_ Context        = (*contextImpl)(nil)
	_ MutableContext = (*contextImpl)(nil)
)

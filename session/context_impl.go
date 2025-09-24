package session

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

type contextImpl struct {
	mu sync.RWMutex

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
}

func newContextImpl(optFns ...Option) *contextImpl {
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

// ----------------------------------------
// State Management
// ----------------------------------------

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

func (c *contextImpl) Client() *http.Client    { return c.opts.Client }
func (c *contextImpl) SetIMEI(imei string)     { c.mu.Lock(); c.imei = imei; c.mu.Unlock() }
func (c *contextImpl) SetUserAgent(ua string)  { c.mu.Lock(); c.userAgent = ua; c.mu.Unlock() }
func (c *contextImpl) SetLanguage(lang string) { c.mu.Lock(); c.language = lang; c.mu.Unlock() }

func (c *contextImpl) SetCookieJar(j http.CookieJar) { c.mu.Lock(); c.cookie = j; c.mu.Unlock() }
func (c *contextImpl) CookieJar() http.CookieJar {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cookie
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

// ----------------------------------------
// Context API
// ----------------------------------------

func (c *contextImpl) UID() string       { return c.uid }
func (c *contextImpl) IMEI() string      { return c.imei }
func (c *contextImpl) UserAgent() string { return c.userAgent }
func (c *contextImpl) Language() string  { return c.language }

func (c *contextImpl) APIType() uint    { return c.apiType }
func (c *contextImpl) APIVersion() uint { return c.apiVersion }

func (c *contextImpl) Options() OptionsSnapshot { return c.opts }
func (c *contextImpl) IsLogging() bool          { return c.opts.Logging }
func (c *contextImpl) LogLevel() uint8          { return c.opts.LogLevel }
func (c *contextImpl) CheckUpdate() bool        { return c.opts.CheckUpdate }

func (c *contextImpl) LoginInfo() *LoginInfo { return c.loginInfo }
func (c *contextImpl) Settings() *Settings   { return c.settings }
func (c *contextImpl) ExtraVer() *ExtraVer   { return c.extraVer }

func (c *contextImpl) SecretKey() string { return c.secretKey }
func (c *contextImpl) Cookies(domains ...string) []*http.Cookie {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cookie == nil {
		return nil
	}

	manager := NewCookieManager(c.cookie)
	return manager.GetCookiesForDomains(domains...)
}

func (c *contextImpl) ZPWServiceMap() *ZpwServiceMap {
	if c.loginInfo == nil {
		return nil
	}
	return &c.loginInfo.ZpwServiceMapV3
}

func (c *contextImpl) ZPWWebsocket() []string {
	if c.loginInfo == nil || len(c.loginInfo.ZpwWebsocket) == 0 {
		return nil
	}
	return c.loginInfo.ZpwWebsocket
}

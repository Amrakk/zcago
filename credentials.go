package zcago

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Amrakk/zcago/internal/errs"
)

type Credentials struct {
	Imei      string      `json:"imei"`
	Cookie    CookieUnion `json:"cookie"`
	UserAgent string      `json:"userAgent"`
	Language  *string     `json:"language,omitempty"`
}

type LoginQROption struct {
	UserAgent *string `json:"userAgent,omitempty"`
	Language  *string `json:"language,omitempty"`
	QRPath    *string `json:"qrPath,omitempty"`
}

type LoginQRCallback func(event any) (any, error)

type SameSite string

const (
	SameSiteDefault SameSite = "default"
	SameSiteLax     SameSite = "lax"
	SameSiteStrict  SameSite = "strict"
	SameSiteNone    SameSite = "none"
)

type Cookie struct {
	Domain         string   `json:"domain"`
	ExpirationDate float32  `json:"expirationDate"`
	HostOnly       bool     `json:"hostOnly"`
	HTTPOnly       bool     `json:"httpOnly"`
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	SameSite       SameSite `json:"sameSite"`
	Secure         bool     `json:"secure"`
	Session        bool     `json:"session"`
	StoreID        *string  `json:"storeId,omitempty"`
	Value          string   `json:"value"`
}

func (c Cookie) ToHTTPCookie() *http.Cookie {
	hc := &http.Cookie{
		Domain:   c.Domain,
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		HttpOnly: c.HTTPOnly,
		Secure:   c.Secure,
	}

	switch c.SameSite {
	case SameSiteStrict:
		hc.SameSite = http.SameSiteStrictMode
	case SameSiteLax:
		hc.SameSite = http.SameSiteLaxMode
	case SameSiteNone:
		hc.SameSite = http.SameSiteNoneMode
	default:
		hc.SameSite = http.SameSiteDefaultMode
	}

	if !c.Session && c.ExpirationDate > 0 {
		hc.Expires = time.Unix(int64(c.ExpirationDate), 0)
	}

	return hc
}

func (c *Cookie) FromHTTPCookie(hc *http.Cookie) {
	c.Domain = hc.Domain
	c.Name = hc.Name
	c.Value = hc.Value
	c.Path = hc.Path
	c.HTTPOnly = hc.HttpOnly
	c.Secure = hc.Secure
	c.HostOnly = false
	c.StoreID = nil

	switch hc.SameSite {
	case http.SameSiteStrictMode:
		c.SameSite = SameSiteStrict
	case http.SameSiteLaxMode:
		c.SameSite = SameSiteLax
	case http.SameSiteNoneMode:
		c.SameSite = SameSiteNone
	default:
		c.SameSite = SameSiteDefault
	}

	if hc.Expires.IsZero() {
		c.Session = true
		c.ExpirationDate = 0
	} else {
		c.Session = false
		c.ExpirationDate = float32(hc.Expires.Unix())
	}
}

type J2Cookie struct {
	URL     string   `json:"url"`
	Cookies []Cookie `json:"cookies"`
}

type CookieUnion struct {
	cookies  []Cookie
	j2cookie *J2Cookie
}

func NewCookieArray(c []Cookie) CookieUnion { return CookieUnion{cookies: c} }
func NewJ2Cookie(j J2Cookie) CookieUnion    { return CookieUnion{j2cookie: &j} }

func (cu CookieUnion) IsValid() bool    { return cu.cookies != nil || cu.j2cookie != nil }
func (cu CookieUnion) IsArray() bool    { return cu.cookies != nil }
func (cu CookieUnion) IsJ2Cookie() bool { return cu.j2cookie != nil }

func (cu CookieUnion) GetCookies() []Cookie {
	if cu.cookies != nil {
		return cu.cookies
	}
	if cu.j2cookie != nil {
		return cu.j2cookie.Cookies
	}
	return nil
}

func (cu CookieUnion) MarshalJSON() ([]byte, error) {
	switch {
	case cu.cookies != nil && cu.j2cookie != nil:
		return nil, errs.NewZCAError("both cookies and j2cookie are set", "CookieUnion.MarshalJSON", nil)
	case cu.cookies != nil:
		return json.Marshal(cu.cookies)
	case cu.j2cookie != nil:
		return json.Marshal(cu.j2cookie)
	default:
		return []byte("null"), nil
	}
}

func (cu *CookieUnion) UnmarshalJSON(b []byte) error {
	trim := bytes.TrimSpace(b)
	if len(trim) == 0 || bytes.Equal(trim, []byte("null")) {
		*cu = CookieUnion{}
		return nil
	}

	if trim[0] == '[' {
		var arr []Cookie
		if err := json.Unmarshal(trim, &arr); err != nil {
			return err
		}
		*cu = CookieUnion{cookies: arr}
		return nil
	}

	var j J2Cookie
	if err := json.Unmarshal(trim, &j); err != nil {
		return err
	}
	*cu = CookieUnion{j2cookie: &j}
	return nil
}

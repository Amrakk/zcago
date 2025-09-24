package session

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
)

var DefaultDomains = []string{"zalo.me", "chat.zalo.me"}

type CookieKey struct {
	Name string
	Path string
}

type CookieManager struct {
	jar http.CookieJar
}

func NewCookieManager(jar http.CookieJar) *CookieManager {
	return &CookieManager{jar: jar}
}

func (cm *CookieManager) GetCookiesForDomains(domains ...string) []*http.Cookie {
	if cm.jar == nil {
		return nil
	}

	targets := domains
	if len(targets) == 0 {
		targets = DefaultDomains
	}

	best := make(map[CookieKey]*http.Cookie)

	for _, host := range targets {
		u := &url.URL{Scheme: "https", Host: host, Path: "/"}
		for _, ck := range cm.jar.Cookies(u) {
			if ck == nil || ck.Name == "" {
				continue
			}

			key := CookieKey{Name: ck.Name, Path: ck.Path}
			cloned := cloneCookie(ck)

			if existing, exists := best[key]; exists {
				best[key] = selectBetterCookie(existing, cloned)
			} else {
				best[key] = cloned
			}
		}
	}

	return sortCookies(best)
}

func cloneCookie(src *http.Cookie) *http.Cookie {
	if src == nil {
		return nil
	}

	return &http.Cookie{
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
}

func selectBetterCookie(a, b *http.Cookie) *http.Cookie {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	rankA := calculateDomainRank(a.Domain)
	rankB := calculateDomainRank(b.Domain)

	if rankA != rankB {
		if rankA < rankB {
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

func calculateDomainRank(domain string) int {
	domain = trimDotAndToLower(domain)
	rank := 0

	for _, part := range strings.Split(domain, ".") {
		if part != "" {
			rank++
		}
	}

	return rank
}

func trimDotAndToLower(s string) string {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, ".") {
		return s[1:]
	}
	return s
}

func sortCookies(cookieMap map[CookieKey]*http.Cookie) []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(cookieMap))
	for _, cookie := range cookieMap {
		cookies = append(cookies, cookie)
	}

	sort.Slice(cookies, func(i, j int) bool {
		domainI := trimDotAndToLower(cookies[i].Domain)
		domainJ := trimDotAndToLower(cookies[j].Domain)

		if domainI != domainJ {
			return domainI < domainJ
		}
		if cookies[i].Path != cookies[j].Path {
			return cookies[i].Path < cookies[j].Path
		}
		return cookies[i].Name < cookies[j].Name
	})

	return cookies
}

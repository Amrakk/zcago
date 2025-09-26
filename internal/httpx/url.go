package httpx

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Amrakk/zcago/session"
)

type URLBuilder struct {
	baseURL string
	params  map[string]string
}

func NewURL(baseURL string) *URLBuilder {
	return &URLBuilder{
		baseURL: baseURL,
		params:  make(map[string]string),
	}
}

func (u *URLBuilder) Param(key, value string) *URLBuilder {
	u.params[key] = value
	return u
}

func (u *URLBuilder) Params(params map[string]string) *URLBuilder {
	for k, v := range params {
		u.params[k] = v
	}
	return u
}

func (u *URLBuilder) Build() string {
	if len(u.params) == 0 {
		return u.baseURL
	}

	parsed, err := url.Parse(u.baseURL)
	if err != nil {
		return u.baseURL
	}

	query := parsed.Query()
	for k, v := range u.params {
		query.Set(k, v)
	}

	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func MakeURL(
	sc session.Context,
	baseURL string,
	params map[string]interface{},
	includeDefaults bool,
) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	query := u.Query()
	for key, value := range params {
		if !query.Has(key) {
			query.Set(key, fmt.Sprintf("%v", value))
		}
	}

	if includeDefaults {
		if !query.Has("zpw_ver") {
			query.Set("zpw_ver", fmt.Sprintf("%v", sc.APIVersion()))
		}
		if !query.Has("zpw_type") {
			query.Set("zpw_type", fmt.Sprintf("%v", sc.APIType()))
		}
	}

	u.RawQuery = query.Encode()
	return u.String()
}

func SignZaloURL(baseURL string, apiType string, params map[string]interface{}) string {
	signKey := GenerateZaloSignKey(apiType, params)

	urlBuilder := NewURL(baseURL)
	for k, v := range params {
		urlBuilder.Param(k, fmt.Sprintf("%v", v))
	}
	urlBuilder.Param("signkey", signKey)

	return urlBuilder.Build()
}

func GenerateZaloSignKey(apiType string, params map[string]interface{}) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	signStr := "zsecure" + apiType
	for _, k := range keys {
		if v := params[k]; v != nil {
			signStr += convertToString(v)
		}
	}

	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}

func convertToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func JoinURL(base, path string) string {
	base = strings.TrimSuffix(base, "/")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		return base
	}

	return base + "/" + path
}

func ResolveURL(base, href string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return href
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return href
	}
	return baseURL.ResolveReference(hrefURL).String()
}

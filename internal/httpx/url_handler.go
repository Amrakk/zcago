package httpx

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"

	"github.com/Amrakk/zcago/session"
)

func MakeURL(
	sc session.Context,
	baseURL string,
	params map[string]interface{},
	includeDefaults bool,
) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
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
	return u.String(), nil
}

func ResolveURL(baseStr, loc string) string {
	locURL, err := url.Parse(loc)
	if err != nil {
		return loc
	}
	if locURL.IsAbs() {
		return locURL.String()
	}
	base, err := url.Parse(baseStr)
	if err != nil {
		return locURL.String()
	}
	return base.ResolveReference(locURL).String()
}

func GetSignKey(apiType string, params map[string]interface{}) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	signStr := "zsecure" + apiType
	for _, k := range keys {
		if v := params[k]; v != nil {
			signStr += toString(v)
		}
	}

	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}

func toString(v interface{}) string {
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

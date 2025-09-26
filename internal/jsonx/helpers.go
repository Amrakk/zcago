package jsonx

import "encoding/json"

func FirstOr[T any](s []T, def T) T {
	if len(s) > 0 {
		return s[0]
	}
	return def
}

func Stringify(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

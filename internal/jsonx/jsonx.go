package jsonx

import "encoding/json"

func Or[T comparable](s, def T) T {
	var zero T
	if s == zero {
		return def
	}
	return s
}

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

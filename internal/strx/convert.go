package strx

import (
	"regexp"
	"strings"
)

var (
	lowerToUpper = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	acronyms     = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
)

func PascalToSnake(s string) string {
	s = lowerToUpper.ReplaceAllString(s, "${1}_${2}")
	s = acronyms.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(s)
}

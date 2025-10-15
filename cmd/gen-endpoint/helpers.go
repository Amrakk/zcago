package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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

func findMarkerAndIndent(src, marker string) (int, string, error) {
	reMarker := regexp.MustCompile(`(?m)^([ \t]*)` + regexp.QuoteMeta(marker) + `\s*$`)
	mLoc := reMarker.FindStringIndex(src)
	if mLoc == nil {
		return 0, "", fmt.Errorf("marker not found: %s", marker)
	}
	indent := ""
	if sm := reMarker.FindStringSubmatch(src[mLoc[0]:mLoc[1]]); len(sm) > 1 {
		indent = sm[1]
	}
	return mLoc[1], indent, nil
}

func readFileString(path string) (string, error) {
	b, err := fs.ReadFile(os.DirFS("."), strings.TrimPrefix(path, "./"))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(b), nil
}

func writeAndFmt(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	// format the file
	if err := exec.Command("gofmt", "-w", path).Run(); err != nil {
		return fmt.Errorf("gofmt %s: %w", path, err)
	}
	return nil
}

func collectNonEmptyLines(region string) []string {
	raw := strings.Split(region, "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "//") {
			continue
		}
		out = append(out, t)
	}
	return out
}

func containsLine(lines []string, line string) bool {
	needle := strings.TrimSpace(line)
	for _, l := range lines {
		if strings.TrimSpace(l) == needle {
			return true
		}
	}
	return false
}

func buildBlock(indent string, lines []string) string {
	var b strings.Builder
	b.WriteString("\n")
	for _, l := range lines {
		b.WriteString(indent)
		b.WriteString(l)
		b.WriteString("\n")
	}
	return b.String()
}

func extractFieldKey(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '('); i >= 0 {
		return s[:i]
	}
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return s[:i]
	}
	return s
}

func extractBindKey(s string) string {
	if m := reBindName.FindStringSubmatch(s); len(m) == 2 {
		return m[1]
	}
	return strings.TrimSpace(s)
}

func toPascal(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toCamel(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

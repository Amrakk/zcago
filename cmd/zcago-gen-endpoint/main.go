package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Amrakk/zcago/internal/strx"
)

type Config struct {
	Name        string // PascalCase
	ApiDir      string
	ApiCore     string
	ApisPath    string
	FactoryName string // camelCase
}

var (
	flagName     = flag.String("name", "", "Endpoint name in PascalCase (e.g., MessagesSendImage)")
	flagApiDir   = flag.String("apiDir", "./api", "Path to the api/ directory")
	flagApiCore  = flag.String("apiCore", "./api/api.go", "Path to api/api.go")
	flagApisPath = flag.String("apisPath", "./apis.go", "Path to apis.go")

	reBindName = regexp.MustCompile(`&a\.e\.([A-Za-z0-9_]+)`)

	markerFields  = "//gen:fields"
	markerBinds   = "//gen:binds"
	markerMethods = "//gen:methods"
)

func main() {
	cfg := parseFlags()
	if err := generateEndpoint(cfg); err != nil {
		fail("Gen-endpoint failed: %v", err)
	}
	fmt.Println("OK.")
}

func parseFlags() *Config {
	flag.Parse()
	if *flagName == "" {
		fail("Error: -name is required")
	}
	name := toPascal(*flagName)
	fac := toCamel(name)

	return &Config{
		Name:        name,
		FactoryName: fac,
		ApiDir:      *flagApiDir,
		ApiCore:     *flagApiCore,
		ApisPath:    *flagApisPath,
	}
}

func generateEndpoint(cfg *Config) error {
	if err := createEndpointFile(cfg); err != nil {
		return err
	}
	if err := patchApiCore(cfg); err != nil {
		return err
	}
	return patchApisFile(cfg)
}

/* ===================== Create endpoint file ===================== */

func createEndpointFile(cfg *Config) error {
	if err := os.MkdirAll(cfg.ApiDir, 0o750); err != nil {
		return fmt.Errorf("creating api dir: %w", err)
	}

	snake := strx.PascalToSnake(cfg.Name)
	dst := filepath.Join(cfg.ApiDir, snake+".go")

	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("endpoint file already exists: %s", dst)
	}
	src := endpointTemplate(cfg.Name)
	return writeAndFmt(dst, src)
}

func endpointTemplate(name string) string {
	fac := toCamel(name)
	return fmt.Sprintf(`package api

import (
	"context"
	"net/http"

	"github.com/Amrakk/zcago/internal/httpx"
	"github.com/Amrakk/zcago/internal/jsonx"
	"github.com/Amrakk/zcago/session"
)

type (
	%sResponse = any
	%sFn       = func(ctx context.Context) (%sResponse, error)
)

func (a *api) %s(ctx context.Context) (%sResponse, error) {
	return a.e.%s(ctx)
}

var %sFactory = apiFactory[%sResponse, %sFn]()(
	func(a *api, sc session.Context, u factoryUtils[%sResponse]) (%sFn, error) {
		base := jsonx.FirstOr(sc.GetZpwService("xxxxxxxxx"), "")
		serviceURL := u.MakeURL(base+"xxxxxxxxx", nil, true)

		return func(ctx context.Context) (%sResponse, error) {
			resp, err := u.Request(ctx, serviceURL, &httpx.RequestOptions{Method: http.MethodGet})
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			return u.Resolve(resp, nil, true)
		}, nil
	},
)
`, name, name, name,
		name, name, name,
		fac, name, name,
		name, name, name,
	)
}

/* ===================== Patch api/api.go ===================== */

func patchApiCore(cfg *Config) error {
	src, err := readFileString(cfg.ApiCore)
	if err != nil {
		return err
	}

	fieldLine := fmt.Sprintf("%s %sFn", cfg.Name, cfg.Name)
	bindLine := fmt.Sprintf("bind(a.ctx, a, &a.e.%s, %sFactory),", cfg.Name, cfg.FactoryName)

	src, err = insertSortedInCurlyBlockAfterMarker(src, markerFields, fieldLine)
	if err != nil {
		return fmt.Errorf("insert field: %w", err)
	}
	src, err = insertSortedInParenBlockAfterMarker(src, markerBinds, bindLine)
	if err != nil {
		return fmt.Errorf("insert bind: %w", err)
	}

	return writeAndFmt(cfg.ApiCore, src)
}

/* ===================== Patch apis.go (interface) ===================== */

func patchApisFile(cfg *Config) error {
	src, err := readFileString(cfg.ApisPath)
	if err != nil {
		return err
	}
	resType := cfg.Name + "Response"
	methodLine := fmt.Sprintf("%s(ctx context.Context) (api.%s, error)", cfg.Name, resType)

	src, err = insertSortedMethodEntriesAfterMarker(src, markerMethods, methodLine)
	if err != nil {
		return fmt.Errorf("insert method: %w", err)
	}
	return writeAndFmt(cfg.ApisPath, src)
}

/* ===================== Sorted insertion (fields) ===================== */

func insertSortedInCurlyBlockAfterMarker(src, marker, line string) (string, error) {
	mEnd, indent, err := findMarkerAndIndent(src, marker)
	if err != nil {
		return "", err
	}
	after := src[mEnd:]
	reClose := regexp.MustCompile(`(?m)^\s*\}\s*$`)
	cLocs := reClose.FindStringIndex(after)
	if cLocs == nil {
		return "", fmt.Errorf("closing '}' not found after marker: %s", marker)
	}
	region := after[:cLocs[0]]
	lines := collectNonEmptyLines(region)

	if containsLine(lines, line) {
		return src, nil
	}
	lines = append(lines, line)
	sort.SliceStable(lines, func(i, j int) bool {
		return strings.ToLower(extractFieldKey(lines[i])) < strings.ToLower(extractFieldKey(lines[j]))
	})
	rebuilt := buildBlock(indent, lines)
	return src[:mEnd] + rebuilt + after[cLocs[0]:], nil
}

/* ===================== Sorted insertion (binds) ===================== */

func insertSortedInParenBlockAfterMarker(src, marker, line string) (string, error) {
	mEnd, indent, err := findMarkerAndIndent(src, marker)
	if err != nil {
		return "", err
	}
	after := src[mEnd:]
	reClose := regexp.MustCompile(`(?m)^\s*\)\s*$`)
	cLocs := reClose.FindStringIndex(after)
	if cLocs == nil {
		return "", fmt.Errorf("closing ')' not found after marker: %s", marker)
	}
	region := after[:cLocs[0]]
	lines := collectNonEmptyLines(region)

	if containsLine(lines, line) {
		return src, nil
	}
	lines = append(lines, line)
	sort.SliceStable(lines, func(i, j int) bool {
		return strings.ToLower(extractBindKey(lines[i])) < strings.ToLower(extractBindKey(lines[j]))
	})
	rebuilt := buildBlock(indent, lines)
	return src[:mEnd] + rebuilt + after[cLocs[0]:], nil
}

/* ===================== Sorted insertion (methods + Godoc preserved) ===================== */

type methodEntry struct {
	docs []string
	code string
	key  string
}

var reClose = regexp.MustCompile(`(?m)^\s*\}\s*$`)

func insertSortedMethodEntriesAfterMarker(src, marker, codeLine string) (string, error) {
	mEnd, indent, err := findMarkerAndIndent(src, marker)
	if err != nil {
		return "", err
	}
	after := src[mEnd:]

	cLocs := reClose.FindStringIndex(after)
	if cLocs == nil {
		return "", fmt.Errorf("closing '}' not found after marker: %s", marker)
	}

	region := after[:cLocs[0]]
	lines := strings.Split(region, "\n")

	entries := make([]methodEntry, 0, len(lines))
	docBuf := []string{}

	for _, raw := range lines {
		t := strings.TrimSpace(raw)

		if strings.HasPrefix(t, "//") {
			docBuf = append(docBuf, raw)
			continue
		}
		if t == "" {
			if len(docBuf) > 0 {
				docBuf = append(docBuf, raw)
			}
			continue
		}

		// code line
		key := strings.ToLower(extractFieldKey(t))
		entries = append(entries, methodEntry{docs: docBuf, code: t, key: key})
		docBuf = nil
	}

	// append if missing
	if !containsMethod(entries, codeLine) {
		t := strings.TrimSpace(codeLine)
		entries = append(entries, methodEntry{
			code: t, key: strings.ToLower(extractFieldKey(t)),
		})
	}

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].key < entries[j].key })

	var b strings.Builder
	b.WriteByte('\n')
	for _, e := range entries {
		for _, dl := range e.docs {
			b.WriteString(dl)
			b.WriteByte('\n')
		}
		b.WriteString(indent)
		b.WriteString(e.code)
		b.WriteByte('\n')
	}

	return src[:mEnd] + b.String() + after[cLocs[0]:], nil
}

func containsMethod(entries []methodEntry, codeLine string) bool {
	needle := strings.TrimSpace(codeLine)
	for _, e := range entries {
		if strings.TrimSpace(e.code) == needle {
			return true
		}
	}
	return false
}

package httpx

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Amrakk/zcago/session"
)

type Level uint8

const (
	Verbose Level = iota
	Debug
	Info
	Warn
	Error
	Success
)

type logChain struct {
	enabled    bool
	minLevel   Level
	out        io.Writer
	withTS     bool
	withColor  bool
	withCaller bool
}

func Logger(sc session.Context) *logChain {
	return &logChain{
		enabled:   sc.IsLogging(),
		minLevel:  Verbose,
		out:       os.Stdout,
		withTS:    true,
		withColor: true,
	}
}

// --- configuration ---

func (c *logChain) SetLevel(l Level) *logChain      { c.minLevel = l; return c }
func (c *logChain) Enable(b bool) *logChain         { c.enabled = b; return c }
func (c *logChain) SetOutput(w io.Writer) *logChain { c.out = w; return c }
func (c *logChain) WithTimestamp(b bool) *logChain  { c.withTS = b; return c }
func (c *logChain) WithColor(b bool) *logChain      { c.withColor = b; return c }
func (c *logChain) WithCaller(b bool) *logChain     { c.withCaller = b; return c }

// --- public API (println-style) ---

func (c *logChain) Verbose(v ...any) *logChain {
	return c.log(Verbose, "VERBOSE", colorMagenta, fmt.Sprint(v...))
}
func (c *logChain) Debug(v ...any) *logChain {
	return c.log(Debug, "DEBUG", colorCyan, fmt.Sprint(v...))
}
func (c *logChain) Info(v ...any) *logChain { return c.log(Info, "INFO", colorBlue, fmt.Sprint(v...)) }
func (c *logChain) Warn(v ...any) *logChain {
	return c.log(Warn, "WARN", colorYellow, fmt.Sprint(v...))
}
func (c *logChain) Error(v ...any) *logChain {
	return c.log(Error, "ERROR", colorRed, fmt.Sprint(v...))
}
func (c *logChain) Success(v ...any) *logChain {
	return c.log(Success, "SUCCESS", colorGreen, fmt.Sprint(v...))
}

// --- public API (printf-style) ---

func (c *logChain) Verbosef(f string, a ...any) *logChain {
	return c.log(Verbose, "VERBOSE", colorMagenta, fmt.Sprintf(f, a...))
}
func (c *logChain) Debugf(f string, a ...any) *logChain {
	return c.log(Debug, "DEBUG", colorCyan, fmt.Sprintf(f, a...))
}
func (c *logChain) Infof(f string, a ...any) *logChain {
	return c.log(Info, "INFO", colorBlue, fmt.Sprintf(f, a...))
}
func (c *logChain) Warnf(f string, a ...any) *logChain {
	return c.log(Warn, "WARN", colorYellow, fmt.Sprintf(f, a...))
}
func (c *logChain) Errorf(f string, a ...any) *logChain {
	return c.log(Error, "ERROR", colorRed, fmt.Sprintf(f, a...))
}
func (c *logChain) Successf(f string, a ...any) *logChain {
	return c.log(Success, "SUCCESS", colorGreen, fmt.Sprintf(f, a...))
}

// --- timestamp-only passthrough (kept for compatibility) ---

func (c *logChain) Timestamp(v ...any) *logChain {
	if !c.enabled {
		return c
	}
	ts := gray(fmtTime(time.Now()))
	fmt.Fprintln(c.out, ts, fmt.Sprint(v...))
	return c
}

func (c *logChain) Timestampf(f string, a ...any) *logChain {
	if !c.enabled {
		return c
	}
	ts := gray(fmtTime(time.Now()))
	fmt.Fprintln(c.out, ts, fmt.Sprintf(f, a...))
	return c
}

// --- core ---

func (c *logChain) log(lvl Level, tag, col string, msg string) *logChain {
	if !c.enabled || lvl < c.minLevel {
		return c
	}

	var b strings.Builder

	// timestamp
	if c.withTS {
		b.WriteString(gray(fmtTime(time.Now())))
		b.WriteByte(' ')
	}

	// level
	if c.withColor {
		b.WriteString(col)
		b.WriteString(tag)
		b.WriteString(colorReset)
	} else {
		b.WriteString(tag)
	}
	b.WriteByte(' ')

	// caller
	if c.withCaller {
		if file, line := caller(3); file != "" {
			b.WriteString(gray(fmt.Sprintf("%s:%d ", file, line)))
		}
	}

	// message
	b.WriteString(msg)

	fmt.Fprintln(c.out, b.String())
	return c
}

// --- helpers ---

const (
	colorReset   = "\x1b[0m"
	colorRed     = "\x1b[31m"
	colorGreen   = "\x1b[32m"
	colorYellow  = "\x1b[33m"
	colorBlue    = "\x1b[34m"
	colorMagenta = "\x1b[35m"
	colorCyan    = "\x1b[36m"
	colorGray    = "\x1b[90m"
)

func gray(s string) string { return colorGray + s + colorReset }

func fmtTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func caller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", 0
	}
	// keep just the tail of the path
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			return file[i+1:], line
		}
	}
	return file, line
}

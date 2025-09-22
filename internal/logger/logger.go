package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
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

type SessionContext interface {
	IsLogging() bool
	LogLevel() uint8
}

type logChain struct {
	enabled  bool
	minLevel Level
}

func Log(sc SessionContext) *logChain {
	return &logChain{
		enabled:  sc.IsLogging(),
		minLevel: Level(sc.LogLevel()),
	}
}

func (c *logChain) SetLevel(l Level) *logChain { c.minLevel = l; return c }
func (c *logChain) Enable(b bool) *logChain    { c.enabled = b; return c }

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

func (c *logChain) log(lvl Level, tag, col string, msg string) *logChain {
	if !c.enabled || lvl < c.minLevel {
		return c
	}

	var b strings.Builder

	// timestamp
	b.WriteString(gray(fmtTime(time.Now())))
	b.WriteByte(' ')

	// level
	b.WriteString(col)
	b.WriteString(tag)
	b.WriteString(colorReset)
	b.WriteByte(' ')

	// message
	b.WriteString(msg)

	fmt.Fprintln(os.Stdout, b.String())
	return c
}

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

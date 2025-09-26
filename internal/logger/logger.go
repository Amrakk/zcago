package logger

import (
	"fmt"
	"os"
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

type Logger struct {
	enabled  bool
	minLevel Level
}

func Log(sc session.Context) *Logger {
	return &Logger{
		enabled:  sc.IsLogging(),
		minLevel: Level(sc.LogLevel()),
	}
}

func (c *Logger) SetLevel(l Level) *Logger { c.minLevel = l; return c }
func (c *Logger) Enable(b bool) *Logger    { c.enabled = b; return c }

func (c *Logger) Verbose(v ...any) *Logger {
	return c.log(Verbose, "VERBOSE", colorMagenta, fmt.Sprint(v...))
}

func (c *Logger) Debug(v ...any) *Logger {
	return c.log(Debug, "DEBUG", colorCyan, fmt.Sprint(v...))
}

func (c *Logger) Info(v ...any) *Logger {
	return c.log(Info, "INFO", colorBlue, fmt.Sprint(v...))
}

func (c *Logger) Warn(v ...any) *Logger {
	return c.log(Warn, "WARN", colorYellow, fmt.Sprint(v...))
}

func (c *Logger) Error(v ...any) *Logger {
	return c.log(Error, "ERROR", colorRed, fmt.Sprint(v...))
}

func (c *Logger) Success(v ...any) *Logger {
	return c.log(Success, "SUCCESS", colorGreen, fmt.Sprint(v...))
}

func (c *Logger) Verbosef(f string, a ...any) *Logger {
	return c.log(Verbose, "VERBOSE", colorMagenta, fmt.Sprintf(f, a...))
}

func (c *Logger) Debugf(f string, a ...any) *Logger {
	return c.log(Debug, "DEBUG", colorCyan, fmt.Sprintf(f, a...))
}

func (c *Logger) Infof(f string, a ...any) *Logger {
	return c.log(Info, "INFO", colorBlue, fmt.Sprintf(f, a...))
}

func (c *Logger) Warnf(f string, a ...any) *Logger {
	return c.log(Warn, "WARN", colorYellow, fmt.Sprintf(f, a...))
}

func (c *Logger) Errorf(f string, a ...any) *Logger {
	return c.log(Error, "ERROR", colorRed, fmt.Sprintf(f, a...))
}

func (c *Logger) Successf(f string, a ...any) *Logger {
	return c.log(Success, "SUCCESS", colorGreen, fmt.Sprintf(f, a...))
}

func (c *Logger) log(lvl Level, tag, col string, msg string) *Logger {
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

	if _, err := fmt.Fprintln(os.Stdout, b.String()); err != nil {
		fmt.Println("logger: failed to write log:", err)
	}
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

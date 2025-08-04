package colorlogger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	reset   = "\033[0m"
	red     = "\033[31;1m"
	yellow  = "\033[33;1m"
	green   = "\033[32;1m"
	cyan    = "\033[36m"
	magenta = "\033[35;1m"
)

func levelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return magenta
	case slog.LevelInfo:
		return green
	case slog.LevelWarn:
		return yellow
	case slog.LevelError:
		return red
	default:
		return reset
	}
}

type ColorHandler struct {
	level slog.Level
}

func NewColorHandler(level slog.Level) slog.Handler {
	return &ColorHandler{level: level}
}

func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder

	color := levelColor(r.Level)
	ts := time.Now().Format("2006-01-02 15:04:05")

	b.WriteString(color)
	b.WriteString(fmt.Sprintf("%s [% -5s] %s", ts, r.Level.String(), r.Message))

	r.Attrs(func(attr slog.Attr) bool {
		b.WriteString(fmt.Sprintf(" %s=%v", attr.Key, attr.Value))
		return true
	})

	b.WriteString(reset)

	_, _ = fmt.Fprintln(os.Stdout, b.String())
	return nil
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return h
}

package dlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

type color int

const (
	timeFormat = "[2006-01-02 15:04:05.251]"

	reset = "\033[0m"

	black        color = 30
	red          color = 31
	green        color = 32
	yellow       color = 33
	blue         color = 34
	magenta      color = 35
	cyan         color = 36
	lightGray    color = 37
	darkGray     color = 90
	lightRed     color = 91
	lightGreen   color = 92
	lightYellow  color = 93
	lightBlue    color = 94
	lightMagenta color = 95
	lightCyan    color = 96
	white        color = 97
)

func colorizer(colorCode color, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(int(colorCode)), v, reset)
}

type DualWriter struct {
	Stdout *os.File
	File   io.Writer
}

func (t DualWriter) Write(p []byte) (n int, err error) {
	n, err = t.WriteStd(p)
	if err != nil {
		return n, err
	}
	n, err = t.WriteFile(p)
	return n, err
}
func (t *DualWriter) WriteStd(p []byte) (n int, err error) {
	n, err = t.Stdout.Write(p)
	if err != nil {
		return n, err
	}
	return n, err
}
func (t *DualWriter) WriteFile(p []byte) (n int, err error) {
	n, err = t.File.Write(p)
	return n, err
}

type Handler struct {
	h        slog.Handler
	r        func([]string, slog.Attr) slog.Attr
	b        *bytes.Buffer
	m        *sync.Mutex
	writer   DualWriter
	colorize bool
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{h: h.h.WithAttrs(attrs), b: h.b, r: h.r, m: h.m, writer: h.writer, colorize: h.colorize}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{h: h.h.WithGroup(name), b: h.b, r: h.r, m: h.m, writer: h.writer, colorize: h.colorize}
}

func (h *Handler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	colorize := func(code color, value string) string {
		return value
	}
	if h.colorize {
		colorize = colorizer
	}

	var level string
	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(r.Level),
	}
	if h.r != nil {
		levelAttr = h.r([]string{}, levelAttr)
	}

	if !levelAttr.Equal(slog.Attr{}) {
		level = levelAttr.Value.String() + ":"

		if r.Level <= slog.LevelDebug {
			level = colorize(lightGray, level)
		} else if r.Level <= slog.LevelInfo {
			level = colorize(cyan, level)
		} else if r.Level < slog.LevelWarn {
			level = colorize(lightBlue, level)
		} else if r.Level < slog.LevelError {
			level = colorize(lightYellow, level)
		} else if r.Level <= slog.LevelError+1 {
			level = colorize(lightRed, level)
		} else if r.Level > slog.LevelError+1 {
			level = colorize(lightMagenta, level)
		}
	}

	var timestamp string
	timeAttr := slog.Attr{
		Key:   slog.TimeKey,
		Value: slog.StringValue(r.Time.Format(timeFormat)),
	}
	if h.r != nil {
		timeAttr = h.r([]string{}, timeAttr)
	}
	if !timeAttr.Equal(slog.Attr{}) {
		timestamp = colorize(lightGray, timeAttr.Value.String())
	}

	var msg string
	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(r.Message),
	}
	if h.r != nil {
		msgAttr = h.r([]string{}, msgAttr)
	}
	if !msgAttr.Equal(slog.Attr{}) {
		msg = colorize(white, msgAttr.Value.String())
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}
	var file string
	if _, ok := attrs["source"].(map[string]interface{}); ok {
		source := attrs["source"].(map[string]interface{})
		if _, ok2 := source["file"]; ok2 {
			line := source["line"]
			source["file"] = source["file"].(string) + ":" + strconv.Itoa(int(line.(float64)))
			file = source["file"].(string)
			delete(attrs, "source")
			attrs["called_function"] = source["function"]
		} else {
			Log.Warn("provided 'source' is overridden may not print source of log")
		}
	} else {
		Log.Warn("provided 'source' is overridden may not print source of log")
	}

	jsonBytes, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
		out.WriteString(" ")
	}
	if len(file) > 0 {
		out.WriteString(file)
		out.WriteString(" ")
	}
	if len(msg) > 0 {
		out.WriteString(msg)
		out.WriteString(" ")
	}
	if len(jsonBytes) > 0 {
		out.WriteString(colorize(green, string(jsonBytes)))
	}

	if r.Level <= slog.LevelDebug {
		_, err := h.writer.WriteFile([]byte(out.String() + "\n"))
		if err != nil {
			return err
		}
		return nil
	}

	_, err = io.WriteString(h.writer, out.String()+"\n")
	if err != nil {
		return err
	}

	return nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

func New(handlerOptions *slog.HandlerOptions, options ...Option) *Handler {
	if handlerOptions == nil {
		handlerOptions = &slog.HandlerOptions{}
	}

	buf := &bytes.Buffer{}
	handler := &Handler{
		b: buf,
		h: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level:       handlerOptions.Level,
			AddSource:   handlerOptions.AddSource,
			ReplaceAttr: suppressDefaults(handlerOptions.ReplaceAttr),
		}),
		r: handlerOptions.ReplaceAttr,
		m: &sync.Mutex{},
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

func NewHandler(writer DualWriter, opts *slog.HandlerOptions) *Handler {
	return New(opts, WithDestinationWriter(writer), WithColor())
}

type Option func(h *Handler)

func WithDestinationWriter(writer DualWriter) Option {
	return func(h *Handler) {
		h.writer = writer
	}
}

func WithColor() Option {
	return func(h *Handler) {
		h.colorize = true
	}
}

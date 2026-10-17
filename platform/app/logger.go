package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"youtrack_backend/internal/platform/config"
	platformlogger "youtrack_backend/internal/platform/logger"
)

// newLogger builds a multi-target structured logger for the process:
//
//   - Terminal output: human-friendly, high-contrast colored pretty handler when
//     attached to a TTY (or forced via LOG_COLOR); otherwise compact JSON.
//   - File persistence: newline-delimited JSON for every event (log_dir/app-YYYY-MM-DD.log).
//   - Error file: WARN+ events duplicated to log_dir/error-YYYY-MM-DD.log for rapid triage.
//
// Log files are partitioned daily and automatically rotate at midnight.
// The returned io.Closer flushes (Sync) and closes all opened log files during
// the cleanup phase. Terminal detection uses os.File.Stat() (ModeCharDevice) so
// no external isatty dependency is required.
func newLogger(settings config.LoggerSettings) (*slog.Logger, io.Closer) {
	out := os.Stderr

	level := slog.LevelInfo
	if parsed, ok := platformlogger.ParseLevel(settings.Level); ok {
		level = parsed
	}

	var isTerminal bool
	if stat, err := out.Stat(); err == nil {
		isTerminal = (stat.Mode() & os.ModeCharDevice) != 0
	}

	jsonOpts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, attr.Value.Time().Format(time.RFC3339))
			}
			return attr
		},
	}

	var terminalHandler slog.Handler
	if isTerminal || settings.ForceColor {
		terminalHandler = &prettyHandler{w: out}
	} else {
		terminalHandler = slog.NewJSONHandler(out, jsonOpts)
	}

	handlers := []slog.Handler{terminalHandler}
	var closers []io.Closer

	if settings.ToFile {
		logDir := settings.LogDir
		if logDir == "" {
			logDir = "logs"
		}

		// 1. All events -> app-YYYY-MM-DD.log in JSON format.
		if appWriter, err := newDailyFileWriter(logDir, "app"); err == nil {
			closers = append(closers, appWriter)
			handlers = append(handlers, slog.NewJSONHandler(appWriter, jsonOpts))
		}

		// 2. WARN+ only -> error-YYYY-MM-DD.log in JSON format.
		if errWriter, err := newDailyFileWriter(logDir, "error"); err == nil {
			closers = append(closers, errWriter)
			handlers = append(handlers, &minLevelHandler{
				handler:  slog.NewJSONHandler(errWriter, jsonOpts),
				minLevel: slog.LevelWarn,
			})
		}
	}

	closer := &logCloser{closers: closers}
	if len(handlers) == 1 {
		return slog.New(terminalHandler), closer
	}
	return slog.New(&multiHandler{handlers: handlers}), closer
}

// prettyHandler renders records with high-contrast ANSI colors and a compact,
// single-line layout optimized for developer terminals.
type prettyHandler struct {
	w  io.Writer
	mu sync.Mutex
}

func (h *prettyHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var levelColor string
	switch {
	case r.Level >= slog.LevelError:
		levelColor = "\033[31m" // Red
	case r.Level >= slog.LevelWarn:
		levelColor = "\033[33m" // Yellow
	case r.Level >= slog.LevelInfo:
		levelColor = "\033[32m" // Green
	case r.Level >= slog.LevelDebug:
		levelColor = "\033[36m" // Cyan
	}

	timeStr := fmt.Sprintf("\033[90m%s\033[0m", r.Time.Format("2006-01-02 03:04:05 PM"))
	levelStr := fmt.Sprintf("%s%-5s\033[0m", levelColor, r.Level.String())

	var methodStr string
	var otherAttrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "method" {
			val := strings.ToUpper(fmt.Sprint(a.Value.Any()))
			var mColor string
			switch val {
			case "GET":
				mColor = "\033[34m" // Blue
			case "POST":
				mColor = "\033[32m" // Green
			case "PUT", "PATCH":
				mColor = "\033[33m" // Yellow
			case "DELETE":
				mColor = "\033[31m" // Red
			default:
				mColor = "\033[35m" // Magenta
			}
			methodStr = fmt.Sprintf(" \033[1m%s%-4s\033[0m", mColor, val)
		} else {
			otherAttrs = append(otherAttrs, a)
		}
		return true
	})

	fmt.Fprintf(h.w, "%s %s%s %-30s", timeStr, levelStr, methodStr, r.Message)

	for _, a := range otherAttrs {
		if a.Key == "status" {
			var sCode int
			switch v := a.Value.Any().(type) {
			case int:
				sCode = v
			case int64:
				sCode = int(v)
			}
			if sCode > 0 {
				var sColor string
				switch {
				case sCode >= 500:
					sColor = "\033[1;37;41m" // Bold White on Red Background
				case sCode >= 400:
					sColor = "\033[1;33m" // Bold Yellow
				case sCode >= 300:
					sColor = "\033[36m" // Cyan
				default:
					sColor = "\033[32m" // Green
				}
				fmt.Fprintf(h.w, " %sstatus=%d\033[0m", sColor, sCode)
				continue
			}
		}
		if a.Key == "path" && r.Level >= slog.LevelWarn {
			fmt.Fprintf(h.w, " \033[1;37mpath=%v\033[0m", a.Value.Any())
			continue
		}
		fmt.Fprintf(h.w, " \033[90m%s=\033[0m%v", a.Key, a.Value.Any())
	}

	fmt.Fprint(h.w, "\n")
	return nil
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *prettyHandler) WithGroup(name string) slog.Handler       { return h }

// multiHandler fans each record out to every enabled sub-handler.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// minLevelHandler filters records to only forward those at or above minLevel.
type minLevelHandler struct {
	handler  slog.Handler
	minLevel slog.Level
}

func (h *minLevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel && h.handler.Enabled(ctx, level)
}

func (h *minLevelHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level < h.minLevel {
		return nil
	}
	return h.handler.Handle(ctx, r)
}

func (h *minLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &minLevelHandler{handler: h.handler.WithAttrs(attrs), minLevel: h.minLevel}
}

func (h *minLevelHandler) WithGroup(name string) slog.Handler {
	return &minLevelHandler{handler: h.handler.WithGroup(name), minLevel: h.minLevel}
}

// logCloser flushes and closes all opened log files in the cleanup phase.
type logCloser struct {
	closers []io.Closer
}

func (lc *logCloser) Close() error {
	var firstErr error
	for _, c := range lc.closers {
		if c == nil {
			continue
		}
		if syncer, ok := c.(interface{ Sync() error }); ok {
			_ = syncer.Sync()
		}
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// dailyFileWriter is a thread-safe io.WriteCloser that writes to log files partitioned
// by the current daily date: <prefix>-YYYY-MM-DD.log inside the specified directory.
// When the day rolls over, it seamlessly flushes and closes the previous file and creates/opens
// the new day's file.
type dailyFileWriter struct {
	mu         sync.Mutex
	dir        string
	prefix     string
	currentDay string
	file       *os.File
	now        func() time.Time
}

func newDailyFileWriter(dir, prefix string) (*dailyFileWriter, error) {
	return newDailyFileWriterWithClock(dir, prefix, time.Now)
}

func newDailyFileWriterWithClock(dir, prefix string, nowFn func() time.Time) (*dailyFileWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log directory %q: %w", dir, err)
	}

	if nowFn == nil {
		nowFn = time.Now
	}

	w := &dailyFileWriter{
		dir:    dir,
		prefix: prefix,
		now:    nowFn,
	}

	if err := w.rotateLocked(w.now()); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *dailyFileWriter) rotateLocked(now time.Time) error {
	day := now.Format("2006-01-02")
	if w.file != nil && w.currentDay == day {
		return nil
	}

	if w.file != nil {
		_ = w.file.Sync()
		_ = w.file.Close()
		w.file = nil
	}

	filename := fmt.Sprintf("%s-%s.log", w.prefix, day)
	filePath := filepath.Join(w.dir, filename)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open daily log file %q: %w", filePath, err)
	}

	w.file = f
	w.currentDay = day
	return nil
}

func (w *dailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	day := now.Format("2006-01-02")
	if w.file == nil || w.currentDay != day {
		if err := w.rotateLocked(now); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)
}

func (w *dailyFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

func (w *dailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		_ = w.file.Sync()
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}
